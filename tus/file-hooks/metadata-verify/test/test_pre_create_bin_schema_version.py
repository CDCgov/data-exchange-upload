import unittest

from pre_create_bin import MetaDefinition
from pre_create_bin import checkMetadataAgainstDefinition

class SchemaVersionTestMethods(unittest.TestCase):

    def test_available_schema_version(self):
        metadata = """
        {
            "schema_version":"1.2"
        }
        """
        definitionObj = [
            MetaDefinition(schema_version = "1.2", fields=[])
        ]
        checkMetadataAgainstDefinition(definitionObj, metadata)
    
    def test_unavailable_schema_version(self):
        metadata = """
        {
            "schema_version":"1.1"
        }
        """
        definitionObj = [
            MetaDefinition(schema_version = "1.2", fields=[])
        ]
        with self.assertRaises(Exception) as context:
            checkMetadataAgainstDefinition(definitionObj, metadata)
        
        self.assertIn('not available', str(context.exception))
        

if __name__ == '__main__':
    unittest.main()

