#!/usr/bin/env python
#

import os
import time

def shell(s):
    print '#',s
    return os.system(s)

def open_browser_windows():
    shell("open http://localhost:8000/ &")
    shell("open http://localhost:8080/ &")


retCode = shell("goapp build")
if retCode == 0:
    open_browser_windows()

    shell("cp app.yaml app.yaml.tmp")
    shell("cat secret/local_extensions_app.yaml >> app.yaml")
    shell("goapp serve ./")
    shell("cp app.yaml.tmp app.yaml")
    shell("rm app.yaml.tmp")

