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
                    (
                        f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: (SFTP + FTP)',
                        enc_value,
                        bck_value,
                        par_value,
                        ['SFTP drive', 'FTP drive']
                    )
                )

                # two virtual disks
                options_matrix.append(
                    (
                        f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: (SFTP + FTP), (SFTP + FTP)',
                        enc_value,
                        bck_value,
                        par_value,
                        ['SFTP drive', 'FTP drive', 'FTP drive', 'SFTP drive']
                    )
                )
            else:
                # one option of every disk
                for p_type in DiskTests.provider_options:
                    options_matrix.append(
                        (
                            f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: {p_type}',
                            enc_value,
                            bck_value,
                            par_value,
                            [p_type]
                        )
                    )

                # one option with all disks
                options_matrix.append(
                    (
                        f'encryption: {enc_key}, backup: {bck_key}, partitioner: {par_key}, disks: {[p_type for p_type in DiskTests.provider_options]}',
                        enc_value,
                        bck_value,
                        par_value,
                        DiskTests.provider_options
                    )
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

        upload.clear()

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

        time.sleep(2)

        os.remove(os.path.join(download_path, filename))

        utils.Logger.debug('Removed the file')

    @parameterized.expand(options_matrix)
    def test_file_tests(self, name, encryption, backup, partitioner, disks):
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
    Uncomment the next method to perform manual tests and debugging.
    The tests are as follows:
        0. encryption: on, backup: RAID10, partitioner: balanced, disks: (SFTP + FTP)
        1. encryption: on, backup: RAID10, partitioner: balanced, disks: (SFTP + FTP), (SFTP + FTP)
        2. encryption: on, backup: RAID10, partitioner: priority, disks: (SFTP + FTP)
        3. encryption: on, backup: RAID10, partitioner: priority, disks: (SFTP + FTP), (SFTP + FTP)
        4. encryption: on, backup: RAID10, partitioner: throughput, disks: (SFTP + FTP)
        5. encryption: on, backup: RAID10, partitioner: throughput, disks: (SFTP + FTP), (SFTP + FTP)
        6. encryption: on, backup: off, partitioner: balanced, disks: SFTP drive
        7. encryption: on, backup: off, partitioner: balanced, disks: FTP drive
        8. encryption: on, backup: off, partitioner: balanced, disks: ['SFTP drive', 'FTP drive']
        9. encryption: on, backup: off, partitioner: priority, disks: SFTP drive
        10. encryption: on, backup: off, partitioner: priority, disks: FTP drive
        11. encryption: on, backup: off, partitioner: priority, disks: ['SFTP drive', 'FTP drive']
        12. encryption: on, backup: off, partitioner: throughput, disks: SFTP drive
        13. encryption: on, backup: off, partitioner: throughput, disks: FTP drive
        14. encryption: on, backup: off, partitioner: throughput, disks: ['SFTP drive', 'FTP drive']
        15. encryption: off, backup: RAID10, partitioner: balanced, disks: (SFTP + FTP)
        16. encryption: off, backup: RAID10, partitioner: balanced, disks: (SFTP + FTP), (SFTP + FTP)
        17. encryption: off, backup: RAID10, partitioner: priority, disks: (SFTP + FTP)
        18. encryption: off, backup: RAID10, partitioner: priority, disks: (SFTP + FTP), (SFTP + FTP)
        19. encryption: off, backup: RAID10, partitioner: throughput, disks: (SFTP + FTP)
        20. encryption: off, backup: RAID10, partitioner: throughput, disks: (SFTP + FTP), (SFTP + FTP)
        21. encryption: off, backup: off, partitioner: balanced, disks: SFTP drive
        22. encryption: off, backup: off, partitioner: balanced, disks: FTP drive
        23. encryption: off, backup: off, partitioner: balanced, disks: ['SFTP drive', 'FTP drive']
        24. encryption: off, backup: off, partitioner: priority, disks: SFTP drive
        25. encryption: off, backup: off, partitioner: priority, disks: FTP drive
        26. encryption: off, backup: off, partitioner: priority, disks: ['SFTP drive', 'FTP drive']
        27. encryption: off, backup: off, partitioner: throughput, disks: SFTP drive
        28. encryption: off, backup: off, partitioner: throughput, disks: FTP drive
        29. encryption: off, backup: off, partitioner: throughput, disks: ['SFTP drive', 'FTP drive']
    """
    """
    def test00_file_manual_tests(self):
        params = options_matrix[28]
        name = params[0]
        encryption = params[1]
        backup = params[2]
        partitioner = params[3]
        disks = params[4]

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
        comparison = filecmp.cmp(os.path.join(self.download_path, '16'), os.path.join(os.getcwd(), '16'),
                                 shallow=False)
        self.assertTrue(comparison)

        comparison = filecmp.cmp(os.path.join(self.download_path, '4'), os.path.join(os.getcwd(), '4'),
                                 shallow=False)
        self.assertTrue(comparison)

        utils.Logger.debug('[file_tests] compared the test files')

        # delete the test files
        FileTests.delete_file(self.driver, '16', self.download_path)
        FileTests.delete_file(self.driver, '4', self.download_path)

        utils.Logger.debug('[file_tests] deleted the test files')

        # delete the volume
        VolumeTests.delete_volume(self.driver, volume_name)
    """

