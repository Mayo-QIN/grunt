#!/usr/bin/env python
from time import time
import numpy as np
from sklearn.cluster import KMeans
import argparse
import nibabel as nib
from sklearn.preprocessing import StandardScaler
np.random.seed(42)
def kmeansseg(imageA, imageB,n_clusters,output):
	t0 = time()
	# try:
	imageA_=nib.load(imageA)
	imageAdata=imageA_.get_data()
	affine=imageA_.get_affine()
	imageB_=nib.load(imageA)
	imageBdata=imageB_.get_data()
	imageAdata=imageAdata.astype(np.float)
	imageBdata=imageBdata.astype(np.float)
	imageAdata=(imageAdata-imageAdata.mean())/imageAdata.std()
	imageBdata=(imageBdata-imageBdata.mean())/imageBdata.std()
	dim =2
	original_image = np.zeros((np.shape(imageAdata)[0], np.shape(imageAdata)[1], np.shape(imageAdata)[2],dim ))
	original_image[:,:, :,0] = imageAdata.copy()
	original_image[:,:, :,1] = imageBdata.copy()
	X = np.reshape(original_image, (np.shape(
		original_image)[0] * np.shape(original_image)[1]* np.shape(original_image)[2], dim))
	k_means = KMeans(int(n_clusters),init='k-means++',n_init=10,n_jobs=1)
	X = StandardScaler().fit_transform(X)
	k_means.fit(X)
	k_means_labels = k_means.labels_
	k_means_cluster_centers = k_means.cluster_centers_
	k_means_labels_unique = np.unique(k_means_labels)
	SEGMENTED = np.reshape(k_means_labels, (np.shape(
			original_image)[0], np.shape(original_image)[1],np.shape(original_image)[2]))
	new_image = nib.Nifti1Image((SEGMENTED), affine)
	nib.save(new_image,output)
	print str((time() - t0))
	# except Exception, e: print e
	return 0


def main(argv):
	kmeansseg(argv.imageA,argv.imageB,argv.clusternumber, argv.output)
	return 0

if __name__ == "__main__":
	parser = argparse.ArgumentParser( description='This file will accept as input a T1 post file and will segment the tumor. Will also require to have the T2 file as well as the ATLAS images it needs')
	parser.add_argument ("-a", "--imageA",  help="The input filename for image A(Input)" , required=True)
	parser.add_argument ("-b", "--imageB",  help="The input filename for image B(Input)" , required=True)
	parser.add_argument ("-o", "--output",  help="The input filename for image B(Input)" , required=True)
	parser.add_argument ("-n", "--clusternumber",  help="The number of cluster" , required=True)
	parser.add_argument('--version', action='version', version='%(prog)s 0.1')
	parser.add_argument("-q", "--quiet",
						  action="store_false", dest="verbose",
						  default=True,
						  help="don't print status messages to stdout")
	args = parser.parse_args()
	main(args)