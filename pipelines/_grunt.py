import requests
import os


class job(object):
    def __init__(self,endpoint):
        self.endpoint = endpoint
        self.files = {}
        self.data = {}

    def call(self, data, files):
        resp = requests.post(self.endpoint.url(), files=files, data=data)
        print resp.text
        self.json = resp.json()
        print ( self.json )
        self.uuid = self.json['uuid']

    def url(self):
        return self.endpoint.address + "/rest/job/" + self.json['uuid']
        
    def status(self):
        self.job_status = requests.get(self.url()).json()
        return self.job_status

    def wait(self):
		self.job_status = requests.get(self.endpoint.address + "/rest/job/wait/" + self.json['uuid'])
		return self.job_status

    def save_all(self,directory):
        for k in self.endpoint.outputs.keys():
            self.save_output(k,directory)
    
    def save_output(self,key,directory):
        print "Saving " + key + " to " + directory
        if key in self.data.keys():
            v = self.data[key]
            try:
                r1 = requests.get( self.url() + "/file/" + key )
                print r1
                with open( os.path.join(directory,v), "wb") as code:
                    code.write(r1.content)
                print 'done'
            except Exception, e: print e

    

    
class endpoint(object):
    """
    Manage an endpoint
    """     
    def __init__(self, address, endpoint):
        self.address = address
        self.endpoint = endpoint
        self.json = requests.get(self.url()).json()
    def parameters(self):
        return self.json['parameters']
    def inputs(self):
        return self.json['input_files']
    def outputs(self):
        return self.json['output_files']
    def url(self):
        return self.address+"/rest/service/"+self.endpoint

    def __call__(self,**kwargs):
        j = job(self)
        print (kwargs)
        for k,v in kwargs.iteritems():
            if k in self.parameters():
                j.data[k] = v
            if k in self.inputs():
                j.files[k] = open(v,'rb')
            if k in self.outputs():
                j.data[k] = v
        j.call(j.data, j.files)
        return j


class grunt(object):
	"""
	A class that manages the comunication with the web services offered by Grunt.
	"""
	def __init__(self, adress,param,files,storelocation, service, user='none', password='none'):
		self.adress = adress
		self.service = service
		self.storelocation=storelocation
		self.user = user
		self.password = password
		self.servicecontactlocation=adress+service
		self.param=param
		self.files=files
		for k, v in param.items():
			setattr(self, k, v)
		for k, v in files.items():
			setattr(self, k, v)
   
	def description(self):
		print "I'm a class for this server %s and specifically this %s end point." % (self.adress, self.service)
 
	def submitjob(self):
		r = requests.post(self.servicecontactlocation, files=self.files, data=self.param)
		self.r=r
		return 0

	def waitforcompletion(self):
		robj=self.r
		ConnObject=robj.json()
		status = requests.get(self.adress+'/rest/job/wait/'+ConnObject.get('uuid'))
		self.status =status
		return 0      

	def download(self):
		robj=self.r
		ConnObject=robj.json()
		filespassed=self.param	 	
	 	for k, v in filespassed.iteritems():
	 		print k,v
	 		try:
				r1 = requests.get( self.adress+'/rest/job/'+ConnObject.get('uuid')+'/file/'+k)
				with open( self.storelocation+v, "wb") as code:
					code.write(r1.content)
				print 'done'
			except Exception, e: print e
			
			return 0  


# n4 = webappserver(adress,param,files,storelocation, service)

if __name__ == "__main__":
    e = endpoint("http://localhost:9901", "copy")
    j = e(input="README.md",output="Copy of README.md")
    print j.status()
    j.wait()
    j.save_output("output", "/tmp/")
