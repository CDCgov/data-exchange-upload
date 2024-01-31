import getopt
import logging
import os
import sys

from dotenv import load_dotenv

from proc_stat_controller import ProcStatController

load_dotenv()
logger = logging.getLogger(__name__)


def post_finish(upload_id):
    ps_api_controller = ProcStatController(os.getenv('PS_API_URL'))
    trace_id, parent_span_id = ps_api_controller.get_trace_by_upload_id(upload_id)
    logger.info(f'Got trace for upload {upload_id} with trace ID {trace_id} and parent span ID {parent_span_id}')

    ps_api_controller.stop_span_for_trace(trace_id, parent_span_id)
    logger.info(f'Stopped child span for parent span {parent_span_id} with stage name of dex-upload')


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

    post_finish(tguid)


if __name__ == '__main__':
    main(sys.argv[1:])
