import time
import datetime

from selenium.webdriver.support import expected_conditions as EC
from selenium.common import TimeoutException
from selenium.webdriver.common.by import By
from selenium.webdriver.support.wait import WebDriverWait


class Bcolors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


class Logger:
    """
    debug_level:
        0 - all messages
        1 - only warnings and errors
        2 - only errors
    """
    debug_level = 0

    @classmethod
    def __print__(cls, color, level, msg):
        print(f"[{datetime.datetime.now()}][{level}]{color}{msg}{Bcolors.ENDC}")

    @classmethod
    def debug(cls, msg):
        if cls.debug_level <= 0:
            cls.__print__(f"{Bcolors.BOLD}{Bcolors.OKCYAN}", "DEBUG", msg)

    @classmethod
    def warn(cls, msg):
        if cls.debug_level <= 1:
            cls.__print__(f"{Bcolors.BOLD}{Bcolors.WARNING}", "DEBUG", msg)

    @classmethod
    def error(cls, msg):
        cls.__print__(f"{Bcolors.BOLD}{Bcolors.FAIL}", "DEBUG", msg)


def wait_til_loaded(delay, browser, css_selector):
    try:
        _ = WebDriverWait(browser, delay).until(EC.presence_of_element_located((By.CSS_SELECTOR, css_selector)))
        Logger.debug("[wait_til_loaded] detected that page was successfully loaded")
    except TimeoutException:
        time.sleep(1)
        Logger.debug("[wait_til_loaded] Timeout Exception fired")