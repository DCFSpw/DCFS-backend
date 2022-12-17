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
    encryption_options = {
        'on': -2,
        'off': -1
    }

    backup_options = {
        'RAID10': -2,
        'off': -1
    }

    partitioner_options = {
        'balanced': -3,
        'priority': -2,
        'throughput': -1
    }

    def setUp(self):
        # initiate selenium
        options = Options()

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

        UserTests.login_as_root(self.driver)

    def tearDown(self):
        self.cursor.close()
        self.db.close()
        self.driver.close()

    @staticmethod
    def add_volume_to_db(cursor, db, name):
        cursor.execute(f'SELECT * FROM volumes WHERE name = \'{name}\'')
        disks = cursor.fetchall()

        # do nothing if the volume already exists
        if len(disks) > 0:
            return

        cursor.execute(
            f'INSERT INTO volumes (uuid, name, user_uuid, backup, encryption, file_partition, created_at, deleted_at)'
            f' VALUES (\'3cd81cf1-740a-11ed-a5fa-00ff94a31ba4\', \'{name}\', \'{UserTests.get_root_uuid(cursor)[0]}\', 1, 1, 1, \'2022-12-04 20:30:56.921\', NULL)')
        db.commit()

    @staticmethod
    def delete_volume_from_db(cursor, db, name):
        cursor.execute(f'SELECT * FROM disks JOIN volumes ON disks.volume_uuid = volumes.uuid WHERE volumes.name = \'{name}\'')
        disks = cursor.fetchall()

        # do not delete a volume with assigned disks
        if len(disks) != 0:
            return

        cursor.execute(f'DELETE FROM volumes WHERE name = \'{name}\'')
        db.commit()

    @staticmethod
    def add_volume(driver, name, backup, encryption, partitioner):
        time.sleep(2)

        # go into the volumes tab
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[2].click()
        time.sleep(1)

        # click the 'new volume' button
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable.q-ma-sm')[0].click()
        time.sleep(1)

        form_fields = driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
        form_fields[0].send_keys(name)  # name
        form_fields = driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__inner.relative-position.col.self-stretch')
        time.sleep(1)

        # backup
        form_fields[1].click()
        time.sleep(1)
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[backup].click()
        time.sleep(1)

        # encryption
        form_fields[2].click()
        time.sleep(1)
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[encryption].click()
        time.sleep(1)

        # partitioner
        form_fields[3].click()
        time.sleep(1)
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item__section.column.q-item__section--main.justify-center')[partitioner].click()
        time.sleep(1)

        # click create
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        utils.Logger.debug('[test00_Volume] added a new volume')

        if backup == VolumeTests.backup_options['RAID10']:
            # confirm backup instruction
            buttons = driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--flat.q-btn--rectangle.text-amber.q-btn--actionable.q-focusable.q-hoverable')
            if len(buttons) > 0:
                buttons[0].click()

    @staticmethod
    def delete_volume(driver, volume_name):
        time.sleep(2)

        # go into the volumes tab
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[2].click()
        time.sleep(1)

        volumes = driver.find_elements(by=By.CSS_SELECTOR, value='.q-card__section.q-card__section--vert.col-auto')
        delete_buttons = driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-negative.text-white.q-btn--actionable.q-focusable.q-hoverable.q-ma-sm')
        for index, volume in enumerate(volumes):
            if volume_name == volume.text:
                delete_buttons[index].click()
                time.sleep(1)
                driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[-1].click()
                time.sleep(2)

        utils.Logger.debug('The volume has been successfully deleted')
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
