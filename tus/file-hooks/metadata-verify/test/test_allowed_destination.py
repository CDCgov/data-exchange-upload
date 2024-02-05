import unittest
from unittest.mock import patch

from pre_create_bin import verify_destination_and_event_allowed
from proc_stat_controller import ProcStatController


@patch.object(ProcStatController, 'create_upload_trace', return_value=('test_trace', 'test_span'))
@patch.object(ProcStatController, 'start_span_for_trace', return_value=('test_trace', 'test_child_span_id'))
@patch.object(ProcStatController, 'stop_span_for_trace')
class TestAllowedDestination(unittest.TestCase):
    def test_should_return_filename_when_allowed_combo_provided(self, create_upload_trace_mock,
                                                                start_span_for_trace_mock,
                                                                stop_span_for_trace_mock):
        allowed_dest = 'dextesting'
        allowed_event_type = 'testevent1'

        definition_filename = verify_destination_and_event_allowed(allowed_dest, allowed_event_type)

        self.assertEqual(definition_filename, 'definitions/dextesting_te1_metadata_definition.json')
        self.assertEqual(create_upload_trace_mock.call_count, 0)
        self.assertEqual(start_span_for_trace_mock.call_count, 0)
        self.assertEqual(stop_span_for_trace_mock.call_count, 0)

    def test_should_throw_when_invalid_destination_provided(self, create_upload_trace_mock, start_span_for_trace_mock,
                                                            stop_span_for_trace_mock):
        non_existent_dest = 'does not exist'
        allowed_event_type = 'testevent1'

        with self.assertRaises(Exception) as context:
            verify_destination_and_event_allowed(non_existent_dest, allowed_event_type)

        self.assertIn('Not a recognized combination', str(context.exception))
        self.assertEqual(create_upload_trace_mock.call_count, 1)
        self.assertEqual(start_span_for_trace_mock.call_count, 1)
        self.assertEqual(stop_span_for_trace_mock.call_count, 1)

    def test_should_throw_when_invalid_event_type_provided(self, create_upload_trace_mock, start_span_for_trace_mock,
                                                            stop_span_for_trace_mock):
        non_existent_dest = 'dextesting'
        allowed_event_type = 'does not exist'

        with self.assertRaises(Exception) as context:
            verify_destination_and_event_allowed(non_existent_dest, allowed_event_type)

        self.assertIn('Not a recognized combination', str(context.exception))
        self.assertEqual(create_upload_trace_mock.call_count, 1)
        self.assertEqual(start_span_for_trace_mock.call_count, 1)
        self.assertEqual(stop_span_for_trace_mock.call_count, 1)
