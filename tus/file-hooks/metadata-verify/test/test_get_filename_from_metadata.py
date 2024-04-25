import unittest

from pre_create_bin import get_filename_from_metadata, verify_filename


class TestGetFilenameFromMetadata(unittest.TestCase):
    def test_should_return_filename_for_filename_field(self):
        test_metadata = {
            'filename': 'test'
        }

        filename = get_filename_from_metadata(test_metadata)

        self.assertEqual('test', filename)

    def test_should_return_filename_for_orig_filename_field(self):
        test_metadata = {
            'original_filename': 'test'
        }

        filename = get_filename_from_metadata(test_metadata)

        self.assertEqual('test', filename)

    def test_should_return_filename_for_ext_filename_field(self):
        test_metadata = {
            'meta_ext_filename': 'test'
        }

        filename = get_filename_from_metadata(test_metadata)

        self.assertEqual('test', filename)

    def test_should_raise_when_no_valid_filename_field_provided(self):
        with self.assertRaises(Exception) as context:
            get_filename_from_metadata({})

        self.assertIn('No filename provided', str(context.exception))

class TestVerifyFilename(unittest.TestCase):
    def test_filename_with_invalid_chars(self):
        invalid_filenames = ['test/']
        for filename in invalid_filenames:
            with self.subTest(filename=filename):
                with self.assertRaises(ValueError):
                    verify_filename(filename)

    def test_filename_without_invalid_chars(self):
        valid_filename = 'valid_filename'
        verify_filename(valid_filename)

if __name__ == '__main__':
    unittest.main()