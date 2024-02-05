import unittest

from pre_create_bin import MetaDefinition
from pre_create_bin import checkMetadataAgainstDefinition
from pre_create_bin import getSchemaVersionToUse


class TestSchemaVersion(unittest.TestCase):
    def test_should_return_oldest_schema_version_when_no_schema_version_provided(self):
        schema_defs = [
            MetaDefinition(schema_version='1.0', fields=[]),
            MetaDefinition(schema_version='1.1', fields=[])
        ]

        selected_schema_def = getSchemaVersionToUse(schema_defs, None)

        self.assertEqual(selected_schema_def.schema_version, '1.0')

    def test_available_schema_version(self):
        metadata = {
            "schema_version": "1.2"
        }

        definitionObj = [
            MetaDefinition(schema_version="1.2", fields=[])
        ]
        checkMetadataAgainstDefinition(definitionObj, metadata)

    def test_unavailable_schema_version(self):
        metadata = {
            "schema_version": "1.1"
        }

        definitionObj = [
            MetaDefinition(schema_version="1.2", fields=[])
        ]
        with self.assertRaises(Exception) as context:
            checkMetadataAgainstDefinition(definitionObj, metadata)

        self.assertIn('not available', str(context.exception))


if __name__ == '__main__':
    unittest.main()
