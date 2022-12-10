import filecmp

from selenium.webdriver.chrome import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium import *
from webdriver_manager.chrome import ChromeDriverManager
from config import Config
from user_test import UserTests
from volume_tests import VolumeTests
from disk_test import DiskTests
from selenium.webdriver import ActionChains

import time
import unittest
import mysql.connector
import utils
import os


class FileTests(unittest.TestCase):
    def setUp(self):
        self.download_path = os.path.join(os.getcwd(), 'downloads')

        # initiate selenium
        options = Options()

        if not os.path.exists(self.download_path):
            os.mkdir(self.download_path)

        prefs = {'download.default_directory': self.download_path}
        options.add_experimental_option('prefs', prefs)

        if Config.headless:
            options.add_argument('--headless')
            options.add_argument('--disable-gpu')

        self.driver = webdriver.Chrome(service=Service(ChromeDriverManager().install()), options=options)
        self.driver.get(Config.frontUrl)
        self.driver.maximize_window()

        # initiate a connection with the DB
        self.db = mysql.connector.connect(
            host=Config.dbAddress,
            database=Config.dbName,
            user=Config.dbUser,
            password=Config.dbPass
        )
        self.cursor = self.db.cursor()

        VolumeTests.add_volume_to_db(self.cursor, self.db, 'test_volume')
        DiskTests.add_remote_disks(self.cursor, self.db, '3cd81cf1-740a-11ed-a5fa-00ff94a31ba4')
        UserTests.login_as_root(self.driver)

    def tearDown(self):
        DiskTests.remove_remote_disks(self.cursor, self.db)
        VolumeTests.delete_volume_from_db(self.cursor, self.db, 'test_volume')
        self.cursor.close()
        self.db.close()
        self.driver.close()

    def test00_File(self):
        """
        uploads, downloads and deletes a 16MB file
        """

        time.sleep(2)

        # expand the volume list
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.row.items-center')[0].click()

        # get the volume options
        volumes = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-manual-focusable')
        for volume in volumes:
            if 'test_volume' in volume.text:
                volume.click()

        # drag and drop the file
        upload = self.driver.find_element('xpath', '//*[@id="fileInput"]')
        file = os.path.join(os.path.abspath(os.getcwd()), '16')
        upload.send_keys(file)

        time.sleep(10)

        # wait til the uploading box disappears
        utils.wait_til_disappeared(1000, self.driver, '.q-card.q-card--dark.q-dark.fixed-bottom-right.upload-progress')

        # temporary solution until the problem with disappearing files is remedied
        time.sleep(10)

        utils.Logger.debug('[test00_File] uploaded the test 16MB file')

        # get the list of files
        files = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-ma-sm.q-pa-md.flex.column.justify-center.items-center.file-block.text-center.full-height.relative-position')
        for file in files:
            if '16' in file.text:
                ac = ActionChains(self.driver)
                ac.context_click(file).perform()

        # download the file
        context_options = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')
        for option in context_options:
            if 'Download' in option.text:
                option.click()

        time.sleep(10)

        # wait til the file downloads
        utils.wait_til_disappeared(1000, self.driver, '.q-card.q-card--dark.q-dark.fixed-bottom-right.upload-progress')

        time.sleep(10)

        utils.Logger.debug('[test00_File] downloaded the 16MB test file')

        comparison = filecmp.cmp(os.path.join(self.download_path, '16'), os.path.join(os.getcwd(), '16'), shallow=False)
        self.assertTrue(comparison)

        utils.Logger.debug('[test00_File] compared the test files')

        # get the list of files
        files = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-ma-sm.q-pa-md.flex.column.justify-center.items-center.file-block.text-center.full-height.relative-position')
        for file in files:
            if '16' in file.text:
                ac = ActionChains(self.driver)
                ac.context_click(file).perform()

        # remove the file
        context_options = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')
        for option in context_options:
            if 'Delete' in option.text:
                option.click()

        # click delete
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()

        time.sleep(5)

        os.remove(os.path.join(self.download_path, '16'))

        utils.Logger.debug('[test00_File] removed the test files')

        self.cursor.execute('DELETE FROM files WHERE name = \'16\'')
        self.db.commit()
        utils.Logger.debug('[test00_File] removed the test file from the db')
