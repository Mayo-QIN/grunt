"""
Execute the following steps before you attempt to run this pipeline

make grunt 
make demo
make ants
make machinelearn
you might need to use sudo (this is attributed to the way docker is designed)

To run the docker webapps use
sudo docker run -d -p 9917:9901 pesscara/machinelearn
sudo docker run -d -p 9916:9901 pesscara/ants

For instance the ports for the webapps offered by pesscara/ants are accessible at port 9916 of you local service 

If you have added custom docker based webapps please initialize. 

This file is depended on the request python library (http://docs.python-requests.org/en/master/user/install/#install). Please see requirements.txt

Example of a curl command 

curl -v -X POST --form clusternumber=6 --form imageA=@T1c.nii.gz --form imageB=@1.nii.gz --form output=cluster.nii.gz ril-gpu10:9913/rest/service/kmeansseg

curl -v -X POST --form fixed=@T1c.nii.gz --form registered=1.nii.gz ril-gpu10:9919/rest/service/n4


"""
from _grunt import grunt

# N4 T1c
adress='http://ril-gpu10:9919'
storelocation="/Users/m112447/Downloads/"
service="/rest/service/n4"
files = {'fixed': open('/Users/m112447/Documents/TestData/T1c.nii.gz', 'rb')}
param = {'registered': 'T1cN4.nii.gz'}
n4 = grunt(adress,param,files,storelocation, service)
n4.submitjob()
n4.jobstatus()
n4.download()
#N4 T2
n4.files = {'fixed': open('/Users/m112447/Documents/TestData/T2.nii.gz', 'rb')}
n4.param = {'registered': 'T2N4.nii.gz'}
n4.submitjob()
n4.jobstatus()
n4.download()
# Register T1 and T2
adress='http://ril-gpu10:9919'
storelocation="/Users/m112447/Downloads/"
service="/rest/service/affine"
files = {'fixed': open('/Users/m112447/Downloads/T1cN4.nii.gz', 'rb'),'moving': open('/Users/m112447/Downloads/T2N4.nii.gz', 'rb')}
param = {'registered': 't2regi.nii.gz'}
regi = grunt(adress,param,files,storelocation, service)
regi.submitjob()
regi.jobstatus()
regi.download()
# Apply clustering 
adress='http://ril-gpu10:9916'
storelocation="/Users/m112447/Downloads/"
service="/rest/service/kmeansseg"
files = {'imageA': open('/Users/m112447/Downloads/T1cN4.nii.gz', 'rb'),'imageB': open('/Users/m112447/Downloads/t2regi.nii.gz', 'rb')}
param = {'output': 'cluster.nii.gz','clusternumber':6}
kmean = grunt(adress,param,files,storelocation, service)
kmean.submitjob()
kmean.jobstatus()
kmean.download()


