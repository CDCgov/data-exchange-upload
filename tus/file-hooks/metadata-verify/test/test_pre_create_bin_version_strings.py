import unittest

from pre_create_bin import get_version_int_from_str


class VersionIntTestMethods(unittest.TestCase):

    def test_versionMajorMinor1(self):
        self.assertEqual(get_version_int_from_str("1.02"), 102)

    def test_versionMajorMinor2(self):
        self.assertEqual(get_version_int_from_str("1.2"), 102)

    def test_versions_not_equal(self):
        self.assertNotEqual(get_version_int_from_str("1.2"), get_version_int_from_str("1.20"))

    def test_versionMajorMinorIter1(self):
        self.assertEqual(get_version_int_from_str("1.02.04"), 10204)

    def test_versionMajorMinorIter2(self):
        self.assertEqual(get_version_int_from_str("1.02.4"), 10204)

    def test_version1p02_greater_1p01(self):
        self.assertGreater(get_version_int_from_str("1.02"), get_version_int_from_str("1.01"))

    def test_version1p02p1_greater_1p01p2(self):
        self.assertGreater(get_version_int_from_str("1.02.1"), get_version_int_from_str("1.01.2"))

    def test_whitespace_equal(self):
        self.assertEqual(get_version_int_from_str(" 1. 3 "), get_version_int_from_str("1.3"))


if __name__ == '__main__':
    unittest.main()
