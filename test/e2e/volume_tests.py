from selenium.webdriver.chrome import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium import *
from webdriver_manager.chrome import ChromeDriverManager
from config import Config
from user_test import UserTests

import time
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

        UserTests.loginAsRoot(self.driver)

    def tearDown(self):
        self.cursor.close()
        self.db.close()
        self.driver.close()

    def test00_Volume(self):
        """
        add a volume
        """

        time.sleep(2)

        # go into the volumes tab
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[2].click()
        time.sleep(1)

        # click the 'new volume' button
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable.q-ma-sm')[0].click()
        time.sleep(1)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
        form_fields[0].send_keys('test_volume')  # name
        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__inner.relative-position.col.self-stretch')
        time.sleep(1)

        # no backup
        form_fields[1].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[-1].click()
        time.sleep(1)

        # no encryption
        form_fields[2].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[-1].click()
        time.sleep(1)

        # balanced partitioner
        form_fields[3].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[-3].click()
        time.sleep(1)

        # click create
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        utils.Logger.debug('[test00_Volume] added a new volume')

    def test01_Volume(self):
        """
        update a volume data
        """

        time.sleep(2)

        # go into the volumes tab
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[2].click()
        time.sleep(1)

        # edit the last volume
        buttons = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')
        buttons[len(buttons) - 1].click()
        time.sleep(1)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
        form_fields[0].send_keys('2')  # name
        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__inner.relative-position.col.self-stretch')
        time.sleep(1)

        # no backup
        form_fields[1].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[-2].click()
        time.sleep(1)

        # no encryption
        form_fields[2].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[-2].click()
        time.sleep(1)

        # balanced partitioner
        form_fields[3].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[-2].click()
        time.sleep(1)

        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[-1].click()
        time.sleep(2)
        utils.Logger.debug('changed the volume data')

    def test02_Volume(self):
        """
        validate the changes made to the newly created volume
        """

        time.sleep(2)

        # go into the volumes tab
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[2].click()
        time.sleep(1)

        # edit the last volume
        buttons = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')
        buttons[len(buttons) - 1].click()
        time.sleep(1)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__inner.relative-position.col.self-stretch')
        self.assertTrue('test_volume2' == self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[0].get_attribute('value'))
        self.assertTrue('RAID1' in form_fields[1].text)
        self.assertTrue('ON' in form_fields[2].text)
        self.assertTrue('Priority' in form_fields[3].text)
        time.sleep(1)

        utils.Logger.debug('[test02_Volume] successfully validated the changed volume data')

    def test02_Volume(self):
        """
        delete the newly created volume
        """

        time.sleep(2)

        # go into the volumes tab
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[2].click()
        time.sleep(1)

        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-negative.text-white.q-btn--actionable.q-focusable.q-hoverable.q-ma-sm')[-1].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[-1].click()
        time.sleep(2)

        self.cursor.execute("SELECT * FROM volumes WHERE name = 'test_volume2' AND deleted_at = NULL")
        result = self.cursor.fetchall()

        self.assertTrue(len(result) == 0)
        utils.Logger.debug('[test02_Volume] the volume has been successfully deleted')
