"""
feature selection demo

"""

from _grunt import *

g = grunt("http://ril-gpu10:9916")
# Syntax 1
j=g.featsel
j.datset="/Users/m112447/Documents/TestData/diab.csv"
j.output="featsel.zip"
job =j()
job.wait()
# Write some output
job.save_output("featsel", "/Users/m112447/Downloads/")

