import json

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
    provider_options = ['SFTP drive', 'FTP drive', 'GoogleDrive']

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

        VolumeTests.add_volume_to_db(self.cursor, self.db, 'test_volume')
        UserTests.login_as_root(self.driver)

    def tearDown(self):
        VolumeTests.delete_volume_from_db(self.cursor, self.db, 'test_volume')
        self.cursor.close()
        self.db.close()
        self.driver.close()

    @staticmethod
    def add_remote_disks(cursor, db, volume_uuid):
        cursor.execute('SELECT uuid FROM providers WHERE name = \'GoogleDrive\'')
        gdrive_provider = cursor.fetchone()[0]

        cursor.execute('SELECT uuid FROM providers WHERE name = \'OneDrive\'')
        onedrive_provider = cursor.fetchone()[0]

        cursor.execute('SELECT uuid FROM providers WHERE name = \'SFTP drive\'')
        sftp_provider = cursor.fetchone()[0]

        """

        # insert a google drive disk
        cursor.execute(f'INSERT INTO disks (uuid, user_uuid, volume_uuid, provider_uuid, credentials, name, created_at, used_space, total_space) VALUES '
                       f'('
                       f'\'922bb881-a419-49cc-b269-298ab88e90bf\','
                       f'\'{UserTests.get_root_uuid(cursor)[0]}\','
                       f'\'{volume_uuid}\','
                       f'\'{gdrive_provider}\','
                       '\'{"accessToken":"ya29.a0AeTM1icyEKlMfnLIr1yUhopNFRFY5im7_IOCBgVl5sVfrRytmXOwO9nRW6IFzSyO1eQrwNcXibYhaI1r4sRmQe_695VM2vnRK1VFeAPOrX46v2iAlBnO9SbiftcUP9JGlXAIejPHNTF1um1Bo1KIHC89VRFMaCgYKARcSARESFQHWtWOmmOqpwZVxMGsFBckDEG9yKw0163","refreshToken":"1//0ckpMy_6Z7b3KCgYIARAAGAwSNwF-L9IrhNPrRl1cDW3hTSKF5W31u9vPt_JHcE5XLMe85qPmHaNA-R8FGwu5hQB9Ka1auue-HVU"}\','
                       f'\'gdrive_test\','
                       f'\'2022-12-04 21:30:56.921\','
                       f'0, 16106127360'
                       f')')
        db.commit()

        # insert a onedrive disk
        cursor.execute(f'INSERT INTO disks (uuid, user_uuid, volume_uuid, provider_uuid, credentials, name, created_at, used_space, total_space) VALUES '
                       f'('
                       f'\'922bb991-a419-49cc-b269-298ab88e90bf\','
                       f'\'{UserTests.get_root_uuid(cursor)[0]}\','
                       f'\'{volume_uuid}\','
                       f'\'{onedrive_provider}\','
                       '\'{"accessToken":"EwBoA8l6BAAUkj1NuJYtTVha+Mogk+HEiPbQo04AAchFRC5PDv5fTwCbjPv/WuQI09Q4Nw4n4OkJsc7NaYnfiC3dT6RkUCUOFdTeegkvpom4UjjIR+SXlkSintNxhBW2giJyTyuXWYrLzip1nz56XRc06i3oKfUMFkY/b7HkZa7KQoiItGV7OqznTv6lUm50qOyhzw7RuHU1sXQ5QSAtVtQlqOYI4O3+vOglcWK+AU6UEytcSbeIpHYbHY+WhEOodClRiTdeqe/IRPcLZeHCe6hkomFNoqJheFtwTpQirzCakNRLPE8uqNz4j4T8YLEeFhlwnQDFaNStVxWXw/V3lQjaWVZ+szJP+NhukBnwEVTiAwEHByKRYW37L+nyB0MDZgAACBWi7LHIguO0OAKCB85qK72YF/9oPRcPAs7hDNx8huTv9cBvwqFSbK9ohYQN/WqMyrwcKpGjtXYxmO9EaTcIFKM0ZWITf62zyQXEp/8p5qvqRL850D3i6f0C94dgGFpKYJaDvpQnU1cbt9d3KDWh3JnQ3P9dDmR2T92ZLXWcKIAa4GnfVknQfULlO4YFvc6MwGZtm57jBpAfZLJ6IlUhuDAHb6xdidcGKJQFh+GdWADQj5/yigBKOCjhUDzoYj0b+Mi+c8hTWOSw+4oBWW0kbmyEQebHF0G9OrIuW9Egm6NAFWvgQQOpalXwyDzUl82xRoNrPz4QKb4gWJrmLrF7PJoVm+V+Vh4MaWkSLtg0hnKZ/Sf4iGlHmD5J2FIyoKq0Q0WB5Khy43USKHsme63VdIzvB61LQ+N7Qqk9Q2RWl1DWYkbBPQLnb/UCf+DZhX2K8fdY6dwiNXya3zB54D+zqv2p0GB+c4jYC0jwhc+J4bYJA0Jt0EBWkoP/sikfD47faDA1C141WYEyT/DC0hx0ws/BZVW4BCxAZTXKR1CVKY+23R4b4tngE0LDcE1YYMViWyllGHSM+YsWr6Kglkqx+QFJaP9WetEp3BI9ocXohsEgKhCpdmE2MHQs/MNCfK7sHbEn54Araq+EZNSXrX8DMdgiaa4veDFEh/lSVike7G3/W/hryRREVLHxW6DatPDbhJBXQYdjYr2kDi8oM7s5/W/CBXRgVfH/0XClKML8bYXRDteJ3hRnBNrWvIc31FUhecvrcAI=","refreshToken":"M.R3_BAY.-CQ7xzIkBvl*6*v0Vrzqq89mqsiHkVGnvJpVUUPoz3rjl76x1JDceujKm6Sef*FNJw47VrBCuj10bxzg7WNfX2hmDSjAqEYRzFxjKa2bH54JejTpP3CPrESkAQg79DMJMDyrxyyCNXbkGB8g6hdtBb7BoTN9tpYL7J2IVh9kwla7D2GxPv36hbxG208!5VK9mNR4N!qbpziol!nVbZBEoPM3xdYFt6ZP02NmCYjmstyoOmTS5VkKjOu9V2!J78z2v9bu0U!*r!JKXL3woLkHpegx8OnMmGJh*9XVk8QOom2Z8"}\','
                       f'\'onedrive_test\','
                       f'\'2022-12-04 21:30:56.921\','
                       f'0, 16106127360'
                       f')')
        db.commit()
        
        """

        # insert an sftp disk
        cursor.execute(f'INSERT INTO disks (uuid, user_uuid, volume_uuid, provider_uuid, credentials, name, created_at, used_space, total_space) VALUES '
                       f'('
                       f'\'922bb551-a419-49cc-b269-298ab88e90bf\','
                       f'\'{UserTests.get_root_uuid(cursor)[0]}\','
                       f'\'{volume_uuid}\','
                       f'\'{sftp_provider}\','
                       '\'{"Login": "dcfs","Password": "UszatekM*01","Host": "34.116.204.182","Port": "2022","Path": "/sftp"}\','
                       f'\'sftp_test\','
                       f'\'2022-12-04 21:30:56.921\','
                       f'0, 16106127360'
                       f')')
        db.commit()

    @staticmethod
    def remove_remote_disks(cursor, db):
        cursor.execute('DELETE FROM disks WHERE uuid = \'922bb551-a419-49cc-b269-298ab88e90bf\'')
        db.commit()

    @staticmethod
    def add_disk(driver, name, volume_name, size, provider):
        time.sleep(1)

        # go into the disks tab
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-focusable.q-hoverable')[3].click()
        time.sleep(1)

        # click the 'new disk' button
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable.q-ma-sm')[0].click()
        time.sleep(1)

        form_fields = driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
        form_fields[0].send_keys(name)  # name
        form_fields = driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__inner.relative-position.col.self-stretch')
        time.sleep(1)

        # choose the volume
        form_fields[1].click()
        time.sleep(1)
        volumes = driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-manual-focusable')
        for volume in volumes:
            if volume_name == volume.text:
                volume.click()
        time.sleep(1)

        # set free space
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[1].send_keys(size)
        time.sleep(1)

        # provider
        form_fields[3].click()
        time.sleep(1)
        options = driver.find_elements(by=By.CSS_SELECTOR, value='.q-item.q-item-type.row.no-wrap.q-item--dark.q-item--clickable.q-link.cursor-pointer.q-manual-focusable')
        for opt in options:
            if provider == opt.text:
                opt.click()
        time.sleep(1)

        if provider == 'SFTP drive' or provider == 'FTP drive':
            with open("./DiskLoginData.json", 'r') as f:
                loginData = json.load(f)

            key = provider.lower()
            form_fields = driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
            form_fields[2].send_keys(loginData[key]['login'])  # login
            form_fields[3].send_keys(loginData[key]['password'])  # password
            form_fields[4].send_keys(loginData[key]['host'])  # host
            form_fields[5].send_keys(loginData[key]['port'])  # port
            form_fields[6].send_keys(loginData[key]['path'])  # path
            time.sleep(1)

        # click create
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-positive.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        time.sleep(1)

        utils.wait_til_disappeared(1000, driver, '.q-field__native.q-placeholder')

        if provider == 'GoogleDrive':
            # login
            driver.find_elements(by=By.CSS_SELECTOR, value='.whsOnd.zHQkBf')[0].send_keys(loginData['GoogleDrive']['login'])
            driver.find_elements(by=By.CSS_SELECTOR, value='.VfPpkd-RLmnJb')[0].click()

            # password
            driver.find_elements(by=By.CSS_SELECTOR, value='.whsOnd.zHQkBf')[0].send_keys(loginData['GoogleDrive']['password'])
            driver.find_elements(by=By.CSS_SELECTOR, value='.VfPpkd-RLmnJb')[0].click()

            # click yes
            driver.find_elements(by=By.CSS_SELECTOR, value='VfPpkd-LgbsSe.VfPpkd-LgbsSe-OWXEXe-dgl2Hf.ksBjEc.lKxP2d.LQeN7.uRo0Xe.TrZEUc.lw1w4b')[0].click()
            driver.find_elements(by=By.CSS_SELECTOR, value='VfPpkd-LgbsSe.VfPpkd-LgbsSe-OWXEXe-INsAgc.VfPpkd-LgbsSe-OWXEXe-dgl2Hf.Rj2Mlf.OLiIxf.PDpWxe.P62QJc.LQeN7.xYnMae.TrZEUc.lw1w4b')[0].click()

        utils.Logger.debug('Added a new volume')

    def test01_Disk(self):
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
