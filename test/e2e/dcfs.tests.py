import selenium
import unittest

import utils
from config import Config
from utils import Logger

from volume_tests import VolumeTests
from user_test import UserTests
from disk_test import DiskTests

Config.set_up()

if __name__ == '__main__':
    Logger.debug_level = 0
    print(utils.Bcolors.ENDC)
    unittest.main(argv=Config.parse_test_args(), warnings='ignore')
