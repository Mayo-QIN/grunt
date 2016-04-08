"""
feature selection demo

"""

from _grunt import *

g = grunt("http://192.168.99.100:9919")
# Syntax 1
j=g.featsel
j.datset="/Users/m112447/Documents/TestData/diab.csv"
j.output="featsel.zip"
job =j()
job.wait()
# Write some output
job.save_output("output", "/Users/m112447/Downloads/")

