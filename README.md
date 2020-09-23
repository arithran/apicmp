# apicmp
A command-line utility that compares **Before** and **After** JSON responses of an API. This tool is ideal for regression testing between Production and QA environments or comparing before and after when rewriting a legacy system using newer technology.


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
   --match value             exact|superset (default: "exact")
   --threads value           10 (default: "4")
   --loglevel value          info (default: "debug")
```

## CSV File 

#### Why?
- CSV files can be imported into any spreadsheet application and be easily edited, sorted & filtered.
- Tools like Splunk & Kibana can easily export log data as CSV. This is useful for regression testing live data.
  - For example: You can search last month's logs on Splunk with a sampling rate of 1:10000, export as CSV,  and regression test that sample data

#### Format
- The CSV file's first line must be a header column (ordering doesn't matter.).
- The following columns have special meaning.
   - `method`: This will be the HTTP Method and will default to `GET` if ommited.
   - `path`: The value is required and is appended to the `--before` & `--after` options provided to the command. Double quotes maybe used if there are any spaces.
   - `body`: This is currently reserved for future use. Currenly this tool only supports the `GET` method.
   - All other fields will forwarded as headers.
  
 Example File:
```
X-User-Id,X-Api-Key,path
0242ac120002,f06f3cf8,/video/2387e4d6a7bede9342150d9afbd0d20f
84198bce24b4,f06f3cf8,/video/3e3a3ecbf14f85db2c74a3b79452f3f1
```

## Examples
```bash
$ apicmp diff \
-B https://api.example.com \
-A https://qa-api.example.com \
-F ~/Downloads/fixture.csv \
-H 'Authorization: Bearer <TOKEN>' \
-H 'Cache-Control: no-cache' \
-I modifiedDate,analytics \
--retry 500 \
--threads 10

```

Output:
> Tip: 'Failed Rows' can be retried with the `--rows` cli option

```bash
$ Summary:
  Total Tests : 273
  Passed      : 263
  Failed      : 10
  Failed Rows : 33,51,102,107,109,152,170,173,239,260
  Time        : 19.990216937s

Issues Found:
       Field       | Issues |          Rows
-------------------+--------+-------------------------
  _http.StatusCode |      6 | 33,102,107,109,239,260
  field1           |      2 |                152,173
  field2           |      2 |                 51,170

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
- [ ] Support DELETE method
- [ ] Support POST, PUT methods


## Contributing
Pull requests are welcome!
