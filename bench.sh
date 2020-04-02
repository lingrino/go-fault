#!/usr/bin/env bash

# This script runs a go benchmark, parses the results, and uploads them to datadog.
# Publishes all three metrics that benchmarks produce (ns/op, bytes/op, allocs/op).
# Metrics will be available in datadog as:
#   go_fault.BENCHMARK_NAME.ns = int
#   go_fault.BENCHMARK_NAME.bytes = int
#   go_fault.BENCHMARK_NAME.allocs = int

# Customize for other repos by changing the metric_namespace variable in prep_env().
# Get more accurate benchmark results by using a longer benchtime duration.
# CPU count is stripped from benchmark name. Recommended you publish using same machine type.
# Run locally by exporting DATADOG_API_KEY. Local runs are tagged with environment:local

# Exit on error
set -e -o pipefail

# Exit if required environment variables are not set.
# Environment variables we depend on should have local values set here.
# https://www.gnu.org/software/bash/manual/html_node/Shell-Parameter-Expansion.html
prep_env() {
    # Required Variables
    ddkey=${DATADOG_API_KEY:?"ERROR: DATADOG_API_KEY is not set"}

    # Static Variables
    benchtime="10s"             # amount of time the benchmark should run for
    metric_namespace="go_fault" # prefix for published metrics
    currenttime="$(date +%s)"   # the current time epoch seconds

    # Optional Variables
    tag_env="${GITHUB_ACTIONS:+"prod"}"
    tag_env="${tag_env:-"local"}"
    tag_sha="${GITHUB_SHA:-"local"}"
    tag_runid="${GITHUB_RUN_ID:-"local"}"
}

# Run a go benchmark and return the results.
run_bench() {
    local bench
    bench="$(go test -run=XXX -bench=. -benchmem -benchtime=${benchtime} -json)"

    # Run the benchmark
    run_bench_output="$bench"
    # Example Output:
    # {"Time":"2020-04-01T13:31:27.531103-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"goos: darwin\n"}
    # {"Time":"2020-04-01T13:31:27.531426-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"goarch: amd64\n"}
    # {"Time":"2020-04-01T13:31:27.531438-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"pkg: github.com/github/go-fault\n"}
    # {"Time":"2020-04-01T13:31:27.531473-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkNoFault\n"}
    # {"Time":"2020-04-01T13:31:27.535617-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkNoFault-8                 \t     674\t      1774 ns/op\t    1705 B/op\t      16 allocs/op\n"}
    # {"Time":"2020-04-01T13:31:27.535633-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkFaultDisabled\n"}
    # {"Time":"2020-04-01T13:31:27.538975-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkFaultDisabled-8           \t     696\t      1960 ns/op\t    1745 B/op\t      17 allocs/op\n"}
    # {"Time":"2020-04-01T13:31:27.538994-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkFaultErrorZeroPercent\n"}
    # {"Time":"2020-04-01T13:31:27.542037-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkFaultErrorZeroPercent-8   \t     621\t      1831 ns/op\t    1746 B/op\t      17 allocs/op\n"}
    # {"Time":"2020-04-01T13:31:27.54205-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkFaultError100Percent\n"}
    # {"Time":"2020-04-01T13:31:27.544146-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"BenchmarkFaultError100Percent-8    \t     282\t      3877 ns/op\t    2299 B/op\t      21 allocs/op\n"}
    # {"Time":"2020-04-01T13:31:27.544167-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"PASS\n"}
    # {"Time":"2020-04-01T13:31:27.54489-05:00","Action":"output","Package":"github.com/github/go-fault","Output":"ok  \tgithub.com/github/go-fault\t0.052s\n"}
    # {"Time":"2020-04-01T13:31:27.544924-05:00","Action":"pass","Package":"github.com/github/go-fault","Elapsed":0.052}
}

# Parse the benchmark results into the output we care about.
parse_bench() {
    local input
    input="$1"

    # Get objects that contain our output strings
    input="$(jq -c 'if (.Action == "output") and (.Output | contains("ns/op")) then . else empty end' <<<"$input")"
    # Trim \n and \t characters from .Output
    input="$(jq '.Output|rtrimstr("\n")|split("\t")|join("")' <<<"$input")"
    # Remove -cpunum from the end of Benchmark
    input="${input//-???/}"
    # Remove leading/trailing quotes
    input="${input//\"/}"
    # Remove whitespace and units
    input="$(awk '{print $1,$3,$5,$7}' <<<"$input")"
    # Output
    parse_bench_output="$input"
    # Example Output:
    # BenchmarkNoFault 1922 1706 16
    # BenchmarkFaultDisabled 1976 1745 17
    # BenchmarkFaultErrorZeroPercent 1621 1745 17
    # BenchmarkFaultError100Percent 3157 1965 20
}

# Publish the benchmark results to datadog
publish() {
    local input
    input="$1"

    # Each line represents a benchmark
    while IFS= read -r line; do
        local bench_name
        bench_name="$(awk '{print $1}' <<<"$line")"

        # Each benchmark has three metrics (ns, bytes, allocs)
        for i in {1..3}; do
            local metric_name
            local metric_value
            metric_name="$(get_metric_name "$i")"
            metric_value="$(awk -v column=$((i + 1)) '{print $column}' <<<"$line")"

            build_json "$bench_name" "$metric_name" "$metric_value"
            publish_metric "$build_json_output"
        done
    done <<<"$input"
}

# Given a number, return the metric name associated with that number.
# A bit of a hack to return the metric name of a metric given that we
# know its position inside a loop.
get_metric_name() {
    local input
    input="$1"

    local metric_name_output
    case "$input" in
    1) metric_name_output="ns" ;;
    2) metric_name_output="bytes" ;;
    3) metric_name_output="allocs" ;;
    *) metric_name_output="ERROR" ;;
    esac

    # Echo so that we can set using mn=$(get_metric_name "$i")
    echo "$metric_name_output"
    # Example Output: ns
}

# Build the JSON object for publishing to datadog
build_json() {
    local bench_name
    local metric_name
    local metric_value
    bench_name="$1"
    metric_name="$2"
    metric_value="$3"

    local name
    name="${metric_namespace}.${bench_name}.${metric_name}"

    local json
    # Static Values
    json={}
    json=$(jq '.series[0].type = "rate"' <<<"$json")
    json=$(jq '.series[0].interval = 1' <<<"$json")
    # Values
    json=$(jq --arg name "${name}" '.series[0].metric = $name' <<<"$json")
    json=$(jq --arg currenttime "${currenttime}" '.series[0].points[0] += [$currenttime]' <<<"$json")
    json=$(jq --arg metric_value "${metric_value}" '.series[0].points[0] += [$metric_value]' <<<"$json")
    # Tags
    json=$(jq --arg env "environment:${tag_env}" '.series[0].tags += [$env]' <<<"$json")
    json=$(jq --arg sha "sha:${tag_sha}" '.series[0].tags += [$sha]' <<<"$json")
    json=$(jq --arg runid "runid:${tag_runid}" '.series[0].tags += [$runid]' <<<"$json")

    build_json_output="$json"
    # Example Output:
    # {
    #   "series": [
    #     {
    #       "type": "rate",
    #       "interval": 1,
    #       "metric": "go_fault.BenchmarkFaultError100Percent.allocs",
    #       "points": [
    #         [
    #           "1585848615",
    #           "20"
    #         ]
    #       ],
    #       "tags": [
    #         "environment:local",
    #         "sha:local",
    #         "runid:local"
    #       ]
    #     }
    #   ]
    # }
}

# Write the metric to datadog
publish_metric() {
    local input
    input="$1"

    local output
    output="$(curl -sS -X POST -H "Content-type: application/json" -d "$input" "https://api.datadoghq.com/api/v1/series?api_key=${ddkey}")"
    printf "%s\n" "$output"
}

prep_env
run_bench
parse_bench "$run_bench_output"
publish "$parse_bench_output"

echo "Success"
