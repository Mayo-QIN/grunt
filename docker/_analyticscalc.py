import numpy as np
import pandas as pd
from sklearn.metrics import roc_curve, auc
from sklearn.metrics import roc_auc_score
import matplotlib
matplotlib.use('pdf')
import matplotlib.pyplot as plt
class color:
   PURPLE = '\033[95m'
   CYAN = '\033[96m'
   DARKCYAN = '\033[36m'
   BLUE = '\033[94m'
   GREEN = '\033[92m'
   YELLOW = '\033[93m'
   RED = '\033[91m'
   BOLD = '\033[1m'
   UNDERLINE = '\033[4m'
   END = '\033[0m'

def analyticscalc(ValuesMetric,tempos,metricDSC):
	print color.UNDERLINE + metricDSC+ color.END
	y_pred =ValuesMetric
	y_true = tempos
	if roc_auc_score(y_true, y_pred)>0.4:
		fpr_full, tpr_full, thresholds_full = roc_curve(tempos, ValuesMetric)
		roc_auc_full = auc(fpr_full, tpr_full)
		maxvalTPFN = (1-fpr_full) + tpr_full
		optimalval=thresholds_full[np.argmax(maxvalTPFN)]
		print color.BOLD + 'treshold, Sensitivity, Specificity' + color.END
		print '{:0.3f}, [{:0.3f} - {:0.3}]'.format(optimalval, tpr_full[np.argmax(maxvalTPFN)],1-fpr_full[np.argmax(maxvalTPFN)] )
		print(color.UNDERLINE + "Original ROC area: {:0.3f}".format(roc_auc_score(y_true, y_pred))+ color.END)
		n_bootstraps = 30
		rng_seed = 42  # control reproducibility
		bootstrapped_scores = []
		rng = np.random.RandomState(rng_seed)
		for i in range(n_bootstraps):
			# bootstrap by sampling with replacement on the prediction indices
			indices = rng.random_integers(0, len(y_pred) - 1, len(y_pred))
			if len(np.unique(y_true[indices])) < 2:
				# We need at least one positive and one negative sample for ROC AUC
				# to be defined: reject the sample
				continue
			score = roc_auc_score(y_true[indices], y_pred[indices])
			bootstrapped_scores.append(score)
			# print("Bootstrap #{} ROC area: {:0.3f}".format(i + 1, score))
		sorted_scores = np.array(bootstrapped_scores)
		sorted_scores.sort()
		confidence_lower = sorted_scores[int(0.05 * len(sorted_scores))]
		confidence_upper = sorted_scores[int(0.95 * len(sorted_scores))]
		print(color.UNDERLINE + "Confidence interval for the score: [{:0.3f} - {:0.3}]".format(
		confidence_lower, confidence_upper)+ color.END)
		print 'Class 1',len(np.nonzero(y_true==1)[0]) ,'Class 2',len(np.nonzero(y_true==0)[0])
		if roc_auc_score(y_true, y_pred)>0.7:
			print color.RED + 'ALARM--------------------------------------------------------'+ color.END
		fpr, tpr, _ = roc_curve(tempos, ValuesMetric)
		return roc_auc_score(y_true, y_pred), optimalval, tpr_full[np.argmax(maxvalTPFN)],1-fpr_full[np.argmax(maxvalTPFN)], confidence_lower, confidence_upper