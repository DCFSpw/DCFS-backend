import argparse
import json
import sys

from utils import Logger


class Config:
    frontUrl = ''
    backUrl = ''
    args = {}
    parameterless_args = ['debug', 'headless', 'no-headless']
    configured = False
    headless = True
    dbAddress = ''
    dbUser = ''
    dbPass = ''
    dbName = ''

    @classmethod
    def set_up(cls):
        if cls.configured:
            return

        parser = argparse.ArgumentParser(description='dcfs.pw end-to-end tests')
        parser.add_argument('--front_url', help='frontend url')
        parser.add_argument('--back_url', help='backend url')
        parser.add_argument('--debug', help="display debug messages", action='store_true')
        parser.add_argument('--headless', help='display browser window', action='store_true')
        parser.add_argument('--db_address', help='mysql server hostname')
        parser.add_argument('--db_user', help='database username')
        parser.add_argument('--db_passwd', help='database password')
        parser.add_argument('--db_name', help='database name')

        cls.args, unknown = parser.parse_known_args()
        cls.args = vars(cls.args)

        if cls.args.get('debug'):
            Logger.debugLevel = 1

        with open("./defaults.json", 'r') as f:
            defaults = json.load(f)

            cls.frontUrl = cls.args['front_url'] if cls.args['front_url'] is not None else defaults['front_url']
            cls.backUrl = cls.args['back_url'] if cls.args['back_url'] is not None else defaults['back_url']
            cls.headless = cls.args['headless'] if cls.args['headless'] is not None else defaults['headless']
            cls.dbAddress = cls.args['db_address'] if cls.args['db_address'] is not None else defaults['database_address']
            cls.dbUser = cls.args['db_user'] if cls.args['db_user'] is not None else defaults['database_user']
            cls.dbPass = cls.args['db_passwd'] if cls.args['db_passwd'] is not None else defaults['database_password']
            cls.dbName = cls.args['db_name'] if cls.args['db_name'] is not None else defaults['database_name']

        Logger.debug(f"args were: {cls.args}")
        Logger.debug(f"Will test backend at: {cls.backUrl} and frontend at {cls.frontUrl}, test will be run {'non-' if not cls.headless else ''}headless")

        cls.configured = True

    @classmethod
    def parse_test_args(cls):
        test_args = sys.argv.copy()

        keys = list(cls.args.keys())
        keys.append('no-headless')

        for key in keys:
            try:
                idx = test_args.index(f"--{key}")
                del test_args[idx]

                if not cls.is_arg_parameterless(key):
                    del test_args[idx]
            except ValueError:
                pass

        print(test_args)
        return test_args

    @classmethod
    def is_arg_parameterless(cls, arg):
        return arg in cls.parameterless_args
