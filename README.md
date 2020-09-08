# apicmp
A command-line utility that compares **Before** and **After** responses of an API. This tool is ideal for regression testing between prod and qa environments.

## Usage
```bash
$ apicmp help diff
NAME:
   apicmp diff - apicmp diff

USAGE:
   apicmp diff [command options] [arguments...]

OPTIONS:
   --before value, -B value  https://api.example.com
   --after value, -A value   https://qa-api.example.com
   --file value, -F value    ~/Downloads/fixtures.csv
   --header value, -H value  'Cache-Control: no-cache'
   --ignore value, -I value  modifiedDate,analytics
   --rows value, -R value    1,7,12 (Rerun failed or specific tests from file)
   --retry value             424,500 (HTTP status codes)
   --match value             full|superset (default: "full")
   --threads value           10 (default: "4")
   --loglevel value          info (default: "debug")
```

## Examples
```bash
$ apicmp diff \
-B https://api.example.com \
-A https://qa-api.example.com \
-F ~/Downloads/fixture.csv \
-H 'Cache-Control: no-cache' \
-I modifiedDate,analytics \
--retry 500 \
--threads 10
```

## Installation
- Download the latest binary for your OS release from the [releases page](https://github.com/arithran/apicmp/releases)
- Rename the file to `apicmp`
- Make the file executable (`chmod +x apicmp`)
- Checkout the help menu for usage instructions `apicmp help`
- (Optional Step) Move it to a folder in your PATH variable. (`mv apicmp /usr/local/bin/`)



## Features
- [x] Print Test Summary
- [x] Trace HTTP Requests and Responses (`--loglevel trace`)
- [x] Multithreading (`--threads 20`)
- [x] Ctrl-C in the middle of a long test run and Print Summary summary of tests that were run


## Contributing
Pull requests are welcome!
