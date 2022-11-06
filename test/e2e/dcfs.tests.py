import selenium
import unittest

from config import Config
from utils import Logger

from volume_tests import VolumeTests

Config.set_up()

if __name__ == '__main__':
    Logger.debug_level = 1
    unittest.main(argv=Config.parse_test_args())
