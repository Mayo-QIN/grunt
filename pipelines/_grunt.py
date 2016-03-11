import requests

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

	def jobstatus(self):
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

