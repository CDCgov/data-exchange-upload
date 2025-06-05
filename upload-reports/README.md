# Application for Upload Data Reporting

This Go application fetches data from the [Processing Status GraphQL API](https://github.com/CDCgov/data-exchange-processing-status), generates a CSV report, and optionally uploads it to an S3 bucket.

## Configuration

### Environment Variables

The application uses the following environment variables that can be set in a `.env` file:

```
PS_API_ENDPOINT=<your_ps_api_url>
S3_BUCKET_NAME=<your_s3_bucket_name>
S3_ENDPOINT=<your_s3_endpoint>
```

If running locally, export these env vars to the shell within which the app will be ran.

#### For Linux/MAC:

```
export PS_API_ENDPOINT=the_ps_api_url
export S3_BUCKET_NAME=the_s3_bucket_name
export S3_ENDPOINT=the_s3_bucket_endpoint
```

#### For Windows:

```
$env:PS_API_ENDPOINT = "the_ps_api_url"
$env:S3_BUCKET_NAME = "the_s3_bucket_name"
$env:S3_ENDPOINT = "the_s3_bucket_endpoint"
```

### Command Line Variables

The application accepts the following command-line variables:

- dataStreams: A comma-separated list of data streams and routes in the format `data-stream-name_route-name`.
- startDate: Start date in UTC (YYYY-MM-DDTHH:MM:SSZ). If not provided, defaults to 24 hours ago from the current time.
- endDate: End date in UTC (YYYY-MM-DDTHH:MM:SSZ). If not provided, defaults to the current time.
- targetEnv: Target environment (default: dev).
- csvOutputPath: Path to save the CSV file (default: current working directory).

## Generating GraphQL Types

This application is using the package [genqlient](https://github.com/Khan/genqlient) to manage type safe graphql implementation.

Before running the application, generate the GraphQL types by executing the following command:

```
go run github.com/Khan/genqlient ./psApi/genqlient.yaml
```

For the above command to work, a current graphql schema must be kept in the `psApi` directory in the file `schema.graphql`. All GraphQL queries that this application needs are located in `upload-reports/psApi/genqlient.graphql`. The above command will generate all needed types based on the queries file and the schema.

## Running the Application

To run the application, use the command:

```
go run ./...
```

## Running the Tests

To run the tests, use the command:

```
go test ./...
```

## GitHub Actions

This application is used by two GitHub Actions located in [data-exchange-upload-devops](https://github.com/cdcent/data-exchange-upload-devops/tree/main/.github/workflows). Reports use the same template `upload-report-template.yml`.
- Action `upload-report-custom.yml` generates a report for a custom range of datastreams, routes, and timeframes.
- Action `upload-report-daily.yml` generates a daily report (24 hours time frame) for a standard set of datastreams.