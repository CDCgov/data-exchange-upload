import getopt
import logging
import os
import sys

from dotenv import load_dotenv

from proc_stat_controller import ProcStatController

from azure.appconfiguration import AzureAppConfigurationClient

connection_string = os.getenv('FEATURE_MANAGER_CONNECTION_STRING')
if connection_string:
    try:
        config_client = AzureAppConfigurationClient.from_connection_string(connection_string)
    except Exception as e:
        raise ValueError(f"Failed to initialize Azure App Configuration: {e}")
else:
    config_client = None

load_dotenv()
logger = logging.getLogger(__name__)

def get_feature_flag(flag_name):
    try:
        fetched_flag = config_client.get_configuration_setting(key=f".appconfig.featureflag/{flag_name}", label=None)
        return fetched_flag.value == "true"
    except Exception as e:
        print(f"Error fetching feature flag {flag_name}: {e}")
        return False

processing_status_reports_enabled = get_feature_flag("PROCESSING_STATUS_REPORTS")
processing_status_traces_enabled = get_feature_flag("PROCESSING_STATUS_TRACES")

def post_finish(upload_id):
    ps_api_controller = ProcStatController(os.getenv('PS_API_URL'))
    span_response = ps_api_controller.get_span_by_upload_id(upload_id, 'dex-upload')
    trace_id = span_response['trace_id']
    span_id = span_response['span_id']
    logger.info(f'Got span for upload {upload_id} with trace ID {trace_id} and span ID {span_id}')

    ps_api_controller.stop_span_for_trace(trace_id, span_id)
    logger.info(f'Stopped child span for parent span {span_id} with stage name of dex-upload')


def main(argv):
    global tguid
    log_level = logging.INFO
    logging.basicConfig(level=log_level)

    opts, args = getopt.getopt(argv, "i:", ["id="])

    for opt, arg in opts:
        if opt in ("-i", "--id"):
            tguid = arg

    if tguid is None:
        raise Exception('No tguid provided')
    if processing_status_traces_enabled:
        post_finish(tguid)


if __name__ == '__main__':
    main(sys.argv[1:])
