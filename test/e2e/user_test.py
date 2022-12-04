import time

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


class UserTests(unittest.TestCase):
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

    def tearDown(self):
        self.cursor.close()
        self.db.close()
        self.driver.close()

    @staticmethod
    def loginAsRoot(driver):
        utils.wait_til_loaded(5, driver, '.q-field__native.q-placeholder')  # wait for the browser to load

        # login as root
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[0].send_keys("root@root.com")
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[1].send_keys("password")
        driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()

        utils.wait_til_loaded(3, driver, '.flex.flex-center.items-center.q-pa-sm')  # wait for the changes in the browser to take place

    def test00_User(self):
        """
        log in as root and log out
        """

        UserTests.loginAsRoot(self.driver)
        self.assertTrue('Distributed Cloud File System' in self.driver.page_source)

        utils.Logger.debug('[test00_User] successfully logged in')

        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-avatar__content.row.flex-center.overflow-hidden')[0].click()
        time.sleep(1)  # wait a second
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        time.sleep(1)

        self.assertTrue('Log in' in self.driver.page_source)
        utils.Logger.debug('[test00_User] successfully logged out')

    def test01_User(self):
        """
        register as a new user
        """

        utils.wait_til_loaded(10, self.driver, '.underline-link')  # wait for the browser to load
        time.sleep(2)

        self.driver.find_elements(by=By.CSS_SELECTOR, value='.underline-link')[0].click()
        time.sleep(2)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')

        form_fields[0].send_keys('First Name')  # first name
        form_fields[1].send_keys('Last Name')  # last name
        form_fields[2].send_keys('email@email.com')  # email
        form_fields[3].send_keys('password')  # password
        form_fields[4].send_keys('password')  # repeat password

        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()
        utils.Logger.debug('[test01_User] successfully registered the new user')

    def test02_User(self):
        """
        login as the newly created user and change the account data
        """

        utils.wait_til_loaded(10, self.driver, '.underline-link')  # wait for the browser to load
        time.sleep(2)

        # login
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[0].send_keys("email@email.com")
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[1].send_keys("password")
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()
        utils.Logger.debug('[test02_User] logged as the newly created user')
        time.sleep(2)

        # enter the account edition panel
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-avatar__content.row.flex-center.overflow-hidden')[0].click()
        time.sleep(1)  # wait a second
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()
        time.sleep(1)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')

        # change the account data
        form_fields[0].send_keys('2')  # change the first name
        form_fields[1].send_keys('2')  # change the last name
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()

        # change the account password
        form_fields[2].send_keys('password')  # enter the current password
        form_fields[3].send_keys('password2')  # enter the new password
        form_fields[4].send_keys('password2')  # repeat the new password
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        utils.Logger.debug('[test02_User] successfully changed the newly created account data')
        time.sleep(2)

        # logout
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-avatar__content.row.flex-center.overflow-hidden')[0].click()
        time.sleep(1)  # wait a second
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        time.sleep(1)
        utils.Logger.debug('[test02_User] successfully logout from the newly created account')

    def test03_User(self):
        """
        validate that the account data was changed successfully and delete the test account
        """

        utils.wait_til_loaded(10, self.driver, '.underline-link')  # wait for the browser to load
        time.sleep(2)

        # login
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[0].send_keys("email@email.com")
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')[1].send_keys("password2")
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()
        utils.Logger.debug('[test03_User] logged as the newly created user with the changed password')
        time.sleep(2)

        # enter the account edition panel
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-avatar__content.row.flex-center.overflow-hidden')[0].click()
        time.sleep(1)  # wait a second
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[0].click()
        time.sleep(1)

        form_fields = self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-field__native.q-placeholder')
        self.assertTrue('First Name2' in form_fields[0].get_attribute('value'))
        self.assertTrue('Last Name2' in form_fields[1].get_attribute('value'))

        # logout
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-avatar__content.row.flex-center.overflow-hidden')[0].click()
        time.sleep(1)  # wait a second
        self.driver.find_elements(by=By.CSS_SELECTOR, value='.q-btn.q-btn-item.non-selectable.no-outline.q-btn--standard.q-btn--rectangle.bg-primary.text-white.q-btn--actionable.q-focusable.q-hoverable')[1].click()
        time.sleep(1)
        utils.Logger.debug('[test03_User] successfully logout from the newly created account')

        self.cursor.execute("DELETE FROM users WHERE first_name = 'First Name2'")
        self.db.commit()
        utils.Logger.debug('[test03_User] deleted the newly created user from the db')
