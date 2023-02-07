import unittest

from pre_create_bin import getVersionIntFromStr

class VersionIntTestMethods(unittest.TestCase):

    def test_json_str1(self):
        testStr1 = """
        [
            {
            "destination_id": "ndlp",
            "ext_events": [
                {
                    "name": "ri",
                    "definition_filename": "ndlp-ri-meta-definition.json"
                },
                {
                    "name": "event1",
                    "definition_filename": null
                }
            ]
            }
        ]
        """
        self.assertEqual(getVersionIntFromStr("1.02"), 102)

if __name__ == '__main__':
    unittest.main()