import unittest

from pre_create_bin import MetaDefinition
from pre_create_bin import get_schema_def_by_version


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

    def test_should_throw_when_requesting_invalid_version(self):
        schema_defs = [
            MetaDefinition(schema_version='1.0', fields=[]),
            MetaDefinition(schema_version='1.1', fields=[])
        ]

        with self.assertRaises(Exception) as context:
            get_schema_def_by_version(schema_defs, 'invalid', {'meta_destination_id': '1234', 'meta_ext_event': '1234'})

        self.assertIn('not available', str(context.exception))


if __name__ == '__main__':
    unittest.main()
