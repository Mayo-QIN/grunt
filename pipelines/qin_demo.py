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
import requests
servername='http://ril-gpu10:9919'
location="/Users/m112447/Downloads/"


# Bias
ServiceContactPoint_1=servername+"/rest/service/n4"
files = {'fixed': open('/Users/m112447/Documents/TestData/T1c.nii.gz', 'rb')}
values = {'registered': 'T1cN4.nii.gz'}
r = requests.post(ServiceContactPoint_1, files=files, data=values)
print dir(r)
print r.json()
ConnObject=r.json()
print ConnObject.get('uuid')
r = requests.get(servername+'/rest/job/wait/'+ConnObject.get('uuid'))
r1 = requests.get( servername+'/rest/job/'+ConnObject.get('uuid')+'/file/registered')
with open( location+values.get('registered'), "wb") as code:
    code.write(r1.content)
print 'Done'
ServiceContactPoint_1=servername+"/rest/service/n4"
files = {'fixed': open('/Users/m112447/Documents/TestData/T2.nii.gz', 'rb')}
values = {'registered': 'T2N4.nii.gz'}
r = requests.post(ServiceContactPoint_1, files=files, data=values)
print dir(r)
print r.json()
ConnObject=r.json()
print ConnObject.get('uuid')
r = requests.get(servername+'/rest/job/wait/'+ConnObject.get('uuid'))
r1 = requests.get( servername+'/rest/job/'+ConnObject.get('uuid')+'/file/registered')
with open( location+values.get('registered'), "wb") as code:
    code.write(r1.content)
print 'Done'
#Register 
ServiceContactPoint_1=servername+"/rest/service/affine"
files = {'fixed': open(location+'/T1cN4.nii.gz', 'rb'),'moving': open(location+'/T2N4.nii.gz', 'rb')}
values = {'registered': 't2regi.nii.gz'}
r = requests.post(ServiceContactPoint_1, files=files, data=values)
r = requests.get(servername+'/rest/job/wait/'+ConnObject.get('uuid'))
r1 = requests.get( servername+'/rest/job/'+ConnObject.get('uuid')+'/file/registered')
with open( location+values.get('registered'), "wb") as code:
    code.write(r1.content)
# Segment
servername='http://ril-gpu10:9916'
ServiceContactPoint_1=servername+"/rest/service/kmeansseg"
files = {'fixed': open(location+'/T1cN4.nii.gz', 'rb'),'moving': open(location+'/t2regi.nii.gz', 'rb')}
values = {'output': 'cluster.nii.gz','clusternumber':6}
r = requests.post(ServiceContactPoint_1, files=files, data=values)
r = requests.get(servername+'/rest/job/wait/'+ConnObject.get('uuid'))
r1 = requests.get( servername+'/rest/job/'+ConnObject.get('uuid')+'/file/output')
with open( location+values.get('registered'), "wb") as code:
    code.write(r1.content)


