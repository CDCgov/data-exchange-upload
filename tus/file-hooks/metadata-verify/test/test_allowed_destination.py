import unittest

from pre_create_bin import verify_destination_and_event_allowed


class TestAllowedDestination(unittest.TestCase):
    def test_should_return_filename_when_allowed_combo_provided(self):
        allowed_dest = 'dextesting'
        allowed_event_type = 'testevent1'

        definition_filename = verify_destination_and_event_allowed(allowed_dest, allowed_event_type)

        self.assertEqual(definition_filename, 'definitions/dextesting_te1_metadata_definition.json')

    def test_should_throw_when_invalid_destination_provided(self):
        non_existent_dest = 'does not exist'
        allowed_event_type = 'testevent1'

        with self.assertRaises(Exception) as context:
            verify_destination_and_event_allowed(non_existent_dest, allowed_event_type)

        self.assertIn('Not a recognized combination', str(context.exception))

    def test_should_throw_when_invalid_event_type_provided(self):
        non_existent_dest = 'dextesting'
        allowed_event_type = 'does not exist'

        with self.assertRaises(Exception) as context:
            verify_destination_and_event_allowed(non_existent_dest, allowed_event_type)

        self.assertIn('Not a recognized combination', str(context.exception))
