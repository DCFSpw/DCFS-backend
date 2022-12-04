from selenium.webdriver.chrome import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium import *
from webdriver_manager.chrome import ChromeDriverManager
from config import Config
from user_test import UserTests
from volume_tests import VolumeTests

import time
import unittest
import mysql.connector
import utils


class DiskTests(unittest.TestCase):
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

        VolumeTests.add_volume_to_db(self.cursor, self.db, 'test_volume')
        UserTests.login_as_root(self.driver)

    def tearDown(self):
        VolumeTests.delete_volume_from_db(self.cursor, self.db, 'test_volume')
        self.cursor.close()
        self.db.close()
        self.driver.close()

    def test00_Disk(self):
        """
        add an SFTP disk
        """

        time.sleep(1)

        # go into the volumes tab
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[3].click()
        time.sleep(1)

        # click the 'new disk' button
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable.q-ma-sm')[0].click()
        time.sleep(1)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
        form_fields[0].send_keys('test_disk')  # name
        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR,value='.q-field__inner.relative-position.col.self-stretch')
        time.sleep(1)

        # the latest volume
        form_fields[1].click()
        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-manual-focusable')[-1].click()
        time.sleep(1)

        # 15 GB of free space
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[1].send_keys('15')
        time.sleep(1)

        # balanced partitioner
        form_fields[3].click()
        time.sleep(1)
        
        options = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-manual-focusable')
        for opt in options:
            if 'SFTP' in opt.text:
                opt.click()
        
        time.sleep(1)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
        form_fields[2].send_keys('server720792_dcfs')  # login
        form_fields[3].send_keys('UszatekM*00')  # password
        form_fields[4].send_keys('ftp.server720792.nazwa.pl')  # host
        form_fields[5].send_keys('22')  # port
        form_fields[6].send_keys('/')  # path
        time.sleep(1)

        # click create
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        time.sleep(10)
        utils.Logger.debug('[test00_Volume] added a new volume')

    def test01_Disk(self):
        """
        validate if the disk has been added
        """

        time.sleep(1)
        self.cursor.execute('SELECT name FROM disks')
        result = self.cursor.fetchall()

        for d in result:
            if d[0] == 'test_disk':
                self.assertTrue(True)

    def test02_Disk(self):
        """
        delete the newly created disk
        """

        time.sleep(1)

        # go into the volumes tab
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[3].click()
        time.sleep(1)

        # fetch all disk names and their delete buttons
        disk_names = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-card__section.q-card__section--vert.col-auto')
        delete_buttons = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-negative.text-white.q-btn--actionable.q-focusable.q-hoverable.q-ma-sm')

        # click the delete button on the test disk
        for idx, label in enumerate(disk_names):
            if 'test_disk' in label.text:
                delete_buttons[idx].click()
                utils.Logger.debug("[test02_Disk] clicked delete on the test_disk")

        time.sleep(1)
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[-1].click()
        time.sleep(2)

        # validate that the test disk has been deleted from the db
        self.db.commit()
        self.cursor.execute('SELECT * FROM disks WHERE name = \'test_disk\'')
        res = self.cursor.fetchall()
        self.assertTrue(len(res) == 0)
