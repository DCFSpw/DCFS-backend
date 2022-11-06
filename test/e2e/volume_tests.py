from selenium.webdriver.chrome import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium import *
from webdriver_manager.chrome import ChromeDriverManager
from config import Config


import unittest
import mysql.connector
import utils


class VolumeTests(unittest.TestCase):
    def setUp(self):
        # initiate selenium
        options = Options()

        if Config.headless:
            options.add_argument('--headless')
            options.add_argument('--disable-gpu')

        self.driver = webdriver.Chrome(service=Service(ChromeDriverManager().install()), options=options)
        self.driver.get(Config.frontUrl)

        # initiate a connection with the DB
        self.db = mysql.connector.connect(
            host=Config.dbAddress,
            database=Config.dbName,
            user=Config.dbUser,
            password=Config.dbPass
        )
        self.cursor = self.db.cursor()

        utils.wait_til_loaded(5, self.driver,
                        '.q-field__native.q-placeholder')  # wait for the browser to load

        # login as root
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[0].send_keys("root@root.com")
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[1].send_keys("password")
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()

        utils.wait_til_loaded(3, self.driver,
                        '.flex.flex-center.items-center.q-pa-sm')  # wait for the changes in the browser to take place

    def tearDown(self):
        self.cursor.close()
        self.db.close()
        self.driver.close()

    def test00_Volume(self):
        self.assertTrue('Distributed Cloud File System' in self.driver.page_source)
        utils.Logger.debug('successfully logged in')
