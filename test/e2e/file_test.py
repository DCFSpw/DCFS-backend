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
from parameterized import parameterized

import time
import unittest
import mysql.connector
import utils
import os
import pytest


options_matrix = []
for enc_index, (enc_key, enc_value) in enumerate(VolumeTests.encryption_options.items()):
    for bkc_index, (bck_key, bck_value) in enumerate(VolumeTests.backup_options.items()):
        for par_index, (par_key, par_value) in enumerate(VolumeTests.partitioner_options.items()):
            if bck_key == 'RAID10':
                # one virtual disk
                options_matrix.append(
                    {
                        "name": f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: (SFTP + FTP)',
                        "encryption": enc_value,
                        "backup": bck_value,
                        "partitioner": par_value,
                        "disks": ['SFTP drive', 'FTP drive']
                    }
                )

                # two virtual disks
                options_matrix.append(
                    {
                        "name": f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: (SFTP + GoogleDrive), (FTP + GoogleDrive)',
                        "encryption": enc_value,
                        "backup": bck_value,
                        "partitioner": par_value,
                        "disks": ['SFTP drive', 'GoogleDrive', 'FTP drive', 'GoogleDrive']
                    }
                )
            else:
                # one option of every disk
                for p_type in DiskTests.provider_options:
                    options_matrix.append(
                        {
                            "name": f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: {p_type}',
                            "encryption": enc_value,
                            "backup": bck_value,
                            "partitioner": par_value,
                            "disks": [p_type]
                        }
                    )

                # one option with all disks
                options_matrix.append(
                    {
                        "name": f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: {[p_type for p_type in DiskTests.provider_options]}',
                        "encryption": enc_value,
                        "backup": bck_value,
                        "partitioner": par_value,
                        "disks": DiskTests.provider_options
                    }
                )


class FileTests(unittest.TestCase):
    def setUp(self):
        Config.set_up()
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

        # VolumeTests.add_volume_to_db(self.cursor, self.db, 'test_volume')
        # DiskTests.add_remote_disks(self.cursor, self.db, '3cd81cf1-740a-11ed-a5fa-00ff94a31ba4')
        UserTests.login_as_root(self.driver)

    def tearDown(self):
        # DiskTests.remove_remote_disks(self.cursor, self.db)
        # VolumeTests.delete_volume_from_db(self.cursor, self.db, 'test_volume')
        self.cursor.close()
        self.db.close()
        self.driver.close()
        
    @staticmethod
    def upload_file(driver, volume_name, filename, download_path):
        time.sleep(1)
        # go into the home tab
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[0].click()

        time.sleep(2)

        # expand the volume list
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.row.items-center')[0].click()

        # get the volume options
        volumes = driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-manual-focusable')
        for volume in volumes:
            if volume_name == volume.text:
                volume.click()
                break

        # drag and drop the file
        upload = driver.find_element('xpath', '//*[@id="fileInput"]')
        file = os.path.join(os.path.abspath(os.getcwd()), filename)
        upload.send_keys(file)

        time.sleep(1)

        # wait til the uploading box disappears
        utils.wait_til_disappeared(1000, driver, '.q-card.q-card--dark.q-dark.fixed-bottom-right.upload-progress')

        utils.Logger.debug('Uploaded the file')

    @staticmethod
    def download_file(driver, filename, download_path):
        time.sleep(1)
        # go into the home tab
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[0].click()

        # get the list of files
        files = driver.find_elements(by=By.CSS_SELECTOR, value='.q-ma-sm.q-pa-md.flex.column.justify-center.items-center.file-block.text-center.full-height.relative-position')
        for file in files:
            if filename in file.text:
                ac = ActionChains(driver)
                ac.context_click(file).perform()

        context_options = driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')
        for option in context_options:
            if 'Download' in option.text:
                option.click()

        time.sleep(2)

        # wait til the file downloads
        utils.wait_til_disappeared(1000, driver, '.q-card.q-card--dark.q-dark.fixed-bottom-right.upload-progress')

        utils.Logger.debug('Downloaded the file')
    @staticmethod
    def delete_file(driver, filename, download_path):
        time.sleep(1)
        # go into the home tab
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[0].click()

        files = driver.find_elements(by=By.CSS_SELECTOR, value='.q-ma-sm.q-pa-md.flex.column.justify-center.items-center.file-block.text-center.full-height.relative-position')
        for file in files:
            if filename in file.text:
                ac = ActionChains(driver)
                ac.context_click(file).perform()

        # remove the file
        context_options = driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')
        for option in context_options:
            if 'Delete' in option.text:
                option.click()

        # click delete
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()

        time.sleep(5)

        os.remove(os.path.join(download_path, filename))

        utils.Logger.debug('Removed the file')

    def test00_file_tests(self):
        params = options_matrix[0]
        name = params["name"]
        encryption = params["encryption"]
        backup = params["backup"]
        partitioner = params["partitioner"]
        disks = params["disks"]

        utils.Logger.debug(f'[file_tests] testing: {name}')
        volume_name = f'{encryption}:{backup}:{partitioner}:{len(disks)}'

        # create volume
        VolumeTests.add_volume(self.driver, volume_name, backup, encryption, partitioner)

        time.sleep(1)

        # add disks
        for disk in disks:
            DiskTests.add_disk(self.driver, disk, volume_name, '5', disk)

        # upload test files
        FileTests.upload_file(self.driver, volume_name, '16', self.download_path)
        FileTests.upload_file(self.driver, volume_name, '4', self.download_path)

        utils.Logger.debug('[file_tests] uploaded the test files')

        # download the files
        FileTests.download_file(self.driver, '16', self.download_path)
        FileTests.download_file(self.driver, '4', self.download_path)

        utils.Logger.debug('[file_tests] downloaded the test files')

        # compare the files
        comparison = filecmp.cmp(os.path.join(self.download_path, '16'), os.path.join(os.getcwd(), '16'), shallow=False)
        self.assertTrue(comparison)

        comparison = filecmp.cmp(os.path.join(self.download_path, '4'), os.path.join(os.getcwd(), '4'), shallow=False)
        self.assertTrue(comparison)

        utils.Logger.debug('[file_tests] compared the test files')

        # delete the test files
        FileTests.delete_file(self.driver, '16', self.download_path)
        FileTests.delete_file(self.driver, '4', self.download_path)

        utils.Logger.debug('[file_tests] deleted the test files')

        # delete the volume
        VolumeTests.delete_volume(self.driver, volume_name)
"""
    @parameterized.expand(options_matrix)
    def test_file_tests(self, params):
        name = params["name"]
        encryption = params["encryption"]
        backup = params["backup"]
        partitioner = params["partitioner"]
        disks = params["disks"]

        utils.Logger.debug(f'[file_tests] testing: {name}')
        volume_name = f'{encryption}:{backup}:{partitioner}:f{len(disks)}'

        # create volume
        VolumeTests.add_volume(self.driver, volume_name, backup, encryption, partitioner)

        # add disks
        for disk in disks:
            DiskTests.add_disk(self.driver, disk, volume_name, '5', disk)

        # upload test files
        FileTests.upload_file(self.driver, volume_name, '16', self.download_path)
        FileTests.upload_file(self.driver, volume_name, '4', self.download_path)

        utils.Logger.debug('[file_tests] uploaded the test files')

        # download the files
        FileTests.download_file(self.driver, '16', self.download_path)
        FileTests.download_file(self.driver, '4', self.download_path)

        utils.Logger.debug('[file_tests] downloaded the test files')

        # compare the files
        comparison = filecmp.cmp(os.path.join(self.download_path, '16'), os.path.join(os.getcwd(), '16'), shallow=False)
        self.assertTrue(comparison)

        comparison = filecmp.cmp(os.path.join(self.download_path, '4'), os.path.join(os.getcwd(), '4'), shallow=False)
        self.assertTrue(comparison)

        utils.Logger.debug('[file_tests] compared the test files')

        # delete the test files
        FileTests.delete_file(self.driver, '16', self.download_path)
        FileTests.delete_file(self.driver, '4', self.download_path)

        utils.Logger.debug('[file_tests] deleted the test files')

        # delete the volume
        VolumeTests.delete_volume(self.driver, volume_name)
        """
