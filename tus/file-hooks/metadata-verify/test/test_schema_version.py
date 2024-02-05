import unittest
from unittest.mock import patch

from pre_create_bin import MetaDefinition
from pre_create_bin import get_schema_def_by_version
from proc_stat_controller import ProcStatController


class TestSchemaVersion(unittest.TestCase):
    def test_should_return_oldest_schema_version_when_no_schema_version_provided(self):
        schema_defs = [
            MetaDefinition(schema_version='1.0', fields=[]),
            MetaDefinition(schema_version='1.1', fields=[])
        ]

        selected_schema_def = get_schema_def_by_version(schema_defs, None, {})

        self.assertEqual(selected_schema_def.schema_version, '1.0')

    def test_should_return_requested_schema_version(self):
        schema_defs = [
            MetaDefinition(schema_version='1.0', fields=[]),
            MetaDefinition(schema_version='1.1', fields=[])
        ]

        selected_schema_def = get_schema_def_by_version(schema_defs, '1.1', {})

        self.assertEqual(selected_schema_def.schema_version, '1.1')

    @patch.object(ProcStatController, 'create_upload_trace', return_value=('test_trace', 'test_span'))
    @patch.object(ProcStatController, 'start_span_for_trace', return_value=('test_trace', 'test_child_span_id'))
    @patch.object(ProcStatController, 'stop_span_for_trace')
    def test_should_throw_when_requesting_invalid_version(self, create_upload_trace_mock, start_span_for_trace_mock,
                                                          stop_span_for_trace_mock):
        schema_defs = [
            MetaDefinition(schema_version='1.0', fields=[]),
            MetaDefinition(schema_version='1.1', fields=[])
        ]

        with self.assertRaises(Exception) as context:
            get_schema_def_by_version(schema_defs, 'invalid', {'meta_destination_id': '1234', 'meta_ext_event': '1234'})

        self.assertIn('not available', str(context.exception))
        self.assertEqual(create_upload_trace_mock.call_count, 1)
        self.assertEqual(start_span_for_trace_mock.call_count, 1)
        self.assertEqual(stop_span_for_trace_mock.call_count, 1)


if __name__ == '__main__':
    unittest.main()
