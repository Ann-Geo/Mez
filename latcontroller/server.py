#!/usr/bin/python


import sys
sys.path.insert(1, '../api/controller')
import grpc

import controller_api_pb2
import controller_api_pb2_grpc
import bisect
import time
import math
import logging
from pandas import DataFrame
from collections import defaultdict
from collections import OrderedDict
from concurrent import futures
import pandas as pd
import io
import numpy as np
import cv2
from PIL import Image
import glob
import time
import os
import shutil
from imutils.video import VideoStream
import argparse
import datetime
import imutils
import subprocess




########################data loading helper funcs#########################

#loads csv file and
#creates hashtables with image size as key and knob and accuracy as values
def createHashtable(csvFile, upperAccLimit, lowerAccLimit):
	df = pd.read_csv(csvFile)
	df = df[df['F1 Score'] <= upperAccLimit]
	df = df[df['F1 Score'] >= lowerAccLimit]
	combos = []
	for i in range(0, df.shape[0]):
		combos.append(int(df.iloc[i][0]))
	
	knobs = []
	with open("all_combos.txt") as fileobj:

		for combo in combos:
			for line in fileobj:
				if "b"+str(combo)+"=" in line:
					knobs.append(line[line.find("=")+2:line.find("\n")-1])

			fileobj.seek(0)
	

	hashTable = defaultdict(list)
	for i in range(0, df.shape[0]):
		size = round(df.iloc[i][1], 2)
		hashTable[size].append(knobs[i])
                hashTable[size].append(round(df.iloc[i][2], 5))
	return hashTable




#loads csv file and calls create hashtable func
def loadData():
	global jaadSimpleHT
	global jaadMediumHT
	global jaadComplexHT

	global dukeSimpleHT
	global dukeMediumHT
	global dukeComplexHT


	jaadSimpleHT = createHashtable("jaad/simple.csv", 0.476886, 0.47)
	jaadSimpleHT = OrderedDict((key, jaadSimpleHT[key]) for key in sorted(jaadSimpleHT))
	jaadMediumHT = createHashtable("jaad/medium.csv", 0.457737, 0.4)
	jaadMediumHT = OrderedDict((key, jaadMediumHT[key]) for key in sorted(jaadMediumHT))
	jaadComplexHT = createHashtable("jaad/complex.csv", 0.41989, 0.35)
	jaadComplexHT = OrderedDict((key, jaadComplexHT[key]) for key in sorted(jaadComplexHT))


	dukeSimpleHT = createHashtable("duke/simple.csv", 0.73315, 0.72)
	dukeSimpleHT = OrderedDict((key, dukeSimpleHT[key]) for key in sorted(dukeSimpleHT))
	dukeMediumHT = createHashtable("duke/medium.csv", 0.35266, 0.34)
	dukeMediumHT = OrderedDict((key, dukeMediumHT[key]) for key in sorted(dukeMediumHT))
	dukeComplexHT = createHashtable("duke/complex.csv", 0.47139, 0.43)
	dukeComplexHT = OrderedDict((key, dukeComplexHT[key]) for key in sorted(dukeComplexHT))






#choose dataset, regime and extract accuracy value from accuracy string
#format = "0.45 jaad complex"
def chooseDataset(accStr):
	global targetAcc
	global targetDataset
	global firstFrame
	global firstFramegray
	global FknobSettings

	targetAcc = accStr.split()[0]
	targetDataset = accStr.split()[1]
	targetRegime = accStr.split()[2]
	if targetDataset == "jaad":
		if targetRegime == "simple":
			targetDataset = jaadSimpleHT
			firstFrame = cv2.imread("jaad/firstframes/00110.png")
			FknobSettings = [0, 1600000, 1740000, 1830000, 1915000]

		elif targetRegime == "medium":
			targetDataset = jaadMediumHT
			firstFrame = cv2.imread("jaad/firstframes/00050.png")
			FknobSettings = [0, 2590000, 2700000 ,2790000, 2870000]
		else:
			targetDataset = jaadComplexHT
			firstFrame = cv2.imread("jaad/firstframes/00230.png")
			FknobSettings = [0, 3350000, 3510000, 3570000, 3645000]

	elif targetDataset == "duke":
		if targetRegime == "simple":
			targetDataset = dukeSimpleHT
			firstFrame = cv2.imread("duke/firstframes/093232.png")
			FknobSettings = [0, 3221000, 3340000, 3410000, 3490000]

		elif targetRegime == "medium":
			targetDataset = dukeMediumHT
			firstFrame = cv2.imread("duke/firstframes/086667.png")
			FknobSettings = [0, 1240000, 1270000, 1310000, 1340000]

		else:
			targetDataset = dukeComplexHT
			firstFrame = cv2.imread("duke/firstframes/071878.png")
			FknobSettings = [0, 1370000, 1430000, 1473000, 1510000]


	firstFrame = imutils.resize(firstFrame, width=500)
        firstFramegray = cv2.cvtColor(firstFrame, cv2.COLOR_BGR2GRAY)
        firstFramegray = cv2.GaussianBlur(firstFramegray, (21, 21), 0)
	
	return targetAcc, targetDataset
		


	
#find image size need to be decreased or increased
def findSizeDelta(currentLat, targetLat):
	global sumLatError

	latDiff = abs(currentLat-targetLat)

	'''
	if currentLat == 0:
		sizeDelta = 0
	elif latDiff < 10:
		sizeDelta = 50000
	elif latDiff >= 10 and latDiff < 20:
		sizeDelta = 100000
	elif latDiff >= 20 and latDiff < 40:
		sizeDelta = 200000
	elif latDiff >= 40 and latDiff < 80:
		sizeDelta = 400000
	elif latDiff >= 80 and latDiff < 100:
		sizeDelta = 600000
	elif latDiff >= 100 and latDiff < 200:
		sizeDelta = 700000
	else:
		sizeDelta = 800000
	
	return sizeDelta

	'''

	sumLatError += latDiff
	if currentLat == 0:
		sizeDelta = 0
	elif latDiff < 10:
		Kp = 3568.9
		Ki = 2446.32
	elif latDiff >= 10 and latDiff < 20:
		Kp = 4523.67
		Ki = 4094.9
	elif latDiff >= 20 and latDiff < 40:
		Kp = 4860.33
		Ki = 2979.11
	elif latDiff >= 40 and latDiff < 80:
		Kp = 3661.5
		Ki = 7189.39
	elif latDiff >= 80 and latDiff < 100:
		Kp = 3546.87
		Ki = 7375.38
	elif latDiff >= 100 and latDiff < 200:
		Kp = 2470.48
		Ki = 6127.73
	else:
		Kp = 2149.77
		Ki = 2807.32

	sizeDelta = Kp*latDiff + Ki*sumLatError
	
	return sizeDelta




#find knobs from hashtable
def findKnobs(newImSize):
	acc = 0.0

	while acc < float(targetAcc):

		ind = bisect.bisect_left(list(targetDataset.keys()), newImSize)
		if ind!=0:
			ind=ind-1	
		newImSize = targetDataset.items()[ind][0]
		knobAndAcc = targetDataset.items()[ind][1]
		knob = knobAndAcc[0]
		acc = knobAndAcc[1]
		knob = [x.strip() for x in knob.split(',')]

	return newImSize, knob, acc
	


#find initial size using the regression model when target latency is given
def findInitialSize(lat):

	#coefficients from regression model
	m = 0.02982
	c = 2.13746
	imSize = (lat - c)/m
	print(imSize)

	return imSize



###################################image modification functions#######################

#apply knobs
def modifyImage(knobs, org_array):
	global prev_frame

	

	image_array = Image.open(io.BytesIO(org_array))

	image_array = np.array(image_array, dtype=np.uint8)
	image_array = apply_frame_differencing(knobs[4], image_array)

	if len(image_array) == 0:
			return ""
	#print(len(image_array))

	prev_frame = image_array

	success, encoded_image = cv2.imencode('.png', image_array)
	image_array = encoded_image.tobytes()

	

	#all modifications - performing one by one
	image_array = change_resolution(knobs[0], org_array)
	image_array = change_colorspace(knobs[1], image_array)
	image_array = change_smoothing_filter_size(knobs[2], image_array)
	image_array = apply_detection_technique(knobs[3], knobs[1], image_array)


	# convert returned numpy array to bytes
	success, encoded_image = cv2.imencode('.png', image_array)
	image_bytes = encoded_image.tobytes()
	return image_bytes



#apply frame diferencing knob
def apply_frame_differencing(f, image_array):

	global prev_frame
	f = f.replace("'", "")

	print(f)

	if (f=='F1'):
		print("f hereeeee")
		dif = FknobSettings[0]
	if (f=='F2'):
		dif = FknobSettings[1]
	if (f=='F3'):
		dif = FknobSettings[2]
	if (f=='F4'):
		dif = FknobSettings[3]
	if (f=='F5'):
		dif = FknobSettings[4]
	
	p_frame_thresh = dif 

	print(type(prev_frame), len(prev_frame))
	print(type(image_array), len(image_array))


	curr_frame = image_array
        diff = cv2.absdiff(curr_frame, prev_frame)
	#print(diff)
	#print("here")
        non_zero_count = np.count_nonzero(diff)
	#print(non_zero_count)
        if non_zero_count > p_frame_thresh:
            	return curr_frame
	else:
		return np.zeros(0)






# change the resolution
def change_resolution(res, image_array):
    res = res.replace("'", "")


    if (res == 'R1'):
        res_image = Image.open(io.BytesIO(image_array))
        return res_image.resize((1920, 1080))

    else:
        if (res == 'R2'):
            width = 1312
            height = 738

        elif (res == 'R3'):
            width = 960
            height = 540
        elif (res == 'R4'):
            width = 640
            height = 360
        else:
            width = 480
            height = 270

        res_image = Image.open(io.BytesIO(image_array))
	res_image.thumbnail((1312, 738))
        return res_image.resize((width, height))


# find the colorspace
def change_colorspace(col, image_array):
    col = col.replace("'", "")


    if (col == 'C1'):
        im_array = np.array(image_array, dtype=np.uint8)
        return im_array

    else:

        if (col == 'C2'):
            col = cv2.COLOR_BGR2GRAY
        elif (col == 'C3'):
            col = cv2.COLOR_BGR2HSV
        elif (col == 'C4'):
            col = cv2.COLOR_BGR2LAB
        else:
            col = cv2.COLOR_BGR2LUV
	

	im_array = np.array(image_array)
        col_image = cv2.cvtColor(im_array, col)
        return col_image



# find the kernel size
def change_smoothing_filter_size(ker, image_array):
    ker = ker.replace("'", "")


    if (ker == 'K1'):
        return image_array

    else:

        if (ker == 'K2'):
            kern = 5
        elif (ker == 'K3'):
            kern = 8
        elif (ker == 'K4'):
            kern = 10
        else:
            kern = 15


        blur_im = cv2.blur(image_array, (kern, kern))
        return blur_im



# code used inside apply_detection_function
def is_contour_bad(c):
    # approximate the contour
    peri = cv2.arcLength(c, True)
    approx = cv2.approxPolyDP(c, 0.02 * peri, True)

    # the contour is 'bad' if it is not a rectangle
    return not len(approx) == 4


# find detection technique
def apply_detection_technique(det, col, image_array):
    det = det.replace("'", "")
    col = col.replace("'", "")


    if (det == 'D1'):
        return image_array

    else:

        if (det == 'D2'):    
            # grab the current frame
            frame = image_array  # cv2.imread(temp_filename)
 

            # resize the frame, convert it to grayscale, and blur it
            frame = imutils.resize(frame, width=500)
	    if (col != 'C2'):
            	gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            	gray = cv2.GaussianBlur(gray, (21, 21), 0)

            else:
            	gray = cv2.GaussianBlur(frame, (21, 21), 0)
            

            # compute the absolute difference between the current frame and
            # first frame
            frameDelta = cv2.absdiff(firstFramegray, gray)
            thresh = cv2.threshold(frameDelta, 25, 255, cv2.THRESH_BINARY)[1]

            # dilate the thresholded image to fill in holes, then find contours
            # on thresholded image
            thresh = cv2.dilate(thresh, None, iterations=2)
            cnts = cv2.findContours(thresh.copy(), cv2.RETR_EXTERNAL,
                                    cv2.CHAIN_APPROX_SIMPLE)
            cnts = cnts[0] if imutils.is_cv2() else cnts[1]

            # loop over the contours
            for c in cnts:

                # compute the bounding box for the contour, draw it on the frame,
                # and update the text
                (x, y, w, h) = cv2.boundingRect(c)
                cv2.rectangle(frame, (x, y), (x + w, y + h), (0, 255, 0), 2)

            return frame



        if (det == 'D3'):
            # define the list of boundaries
            boundaries = [
                ([0, 250, 0], [0, 255, 0])
            ]

	    #colorspace C2 should be dealt separately
	    if (col == 'C2'):
	    	boundaries = [
                	([0, 0, 0], [0, 5, 0])
            	]
            
            # grab the current frame 
            frame = image_array  # cv2.imread(temp_filename)
          

            # resize the frame, convert it to grayscale, and blur it
            frame = imutils.resize(frame, width=500)
	    if (col != 'C2'):
            	gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            	gray = cv2.GaussianBlur(gray, (21, 21), 0)

            else:
            	gray = cv2.GaussianBlur(frame, (21, 21), 0)
           


            # compute the absolute difference between the current frame and
            # first frame
            frameDelta = cv2.absdiff(firstFramegray, gray)
            thresh = cv2.threshold(frameDelta, 25, 255, cv2.THRESH_BINARY)[1]

            # dilate the thresholded image to fill in holes, then find contours
            # on thresholded image
            thresh = cv2.dilate(thresh, None, iterations=2)
            cnts = cv2.findContours(thresh.copy(), cv2.RETR_EXTERNAL,
                                    cv2.CHAIN_APPROX_SIMPLE)
            cnts = cnts[0] if imutils.is_cv2() else cnts[1]

            # loop over the contours
            for c in cnts:

                # compute the bounding box for the contour, draw it on the frame,
                # and update the text
                (x, y, w, h) = cv2.boundingRect(c)
                cv2.rectangle(frame, (x, y), (x + w, y + h), (0, 255, 0), 2)
      

            ###################################  D3 starts #########################################
           
            if (col == 'C2'):
		frame = cv2.merge((frame,frame,frame))

            mask = np.ones(frame.shape[:2], dtype="uint8")

	    if (col != 'C2'):
            	gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            	edged = cv2.Canny(gray, 50, 100)
	    else:
		edged = cv2.Canny(frame, 50, 100)	

	   
            # loop over the boundaries
            for (lower, upper) in boundaries:
                # create NumPy arrays from the boundaries
                lower = np.array(lower, dtype="uint8")
                upper = np.array(upper, dtype="uint8")

                # find the colors within the specified boundaries and apply
                # the mask
                mask1 = cv2.inRange(frame, lower, upper)
                cnts = cv2.findContours(
                    mask1.copy(), cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
                cnts = cnts[0] if imutils.is_cv2() else cnts[1]

                
		if (col != 'C2'):
		    for c in cnts:
                    	# if the contour is bad, draw it on the mask
                    	cv2.drawContours(mask1, [c], -1, 255, -1)
		else:
		    for c in cnts:
			cv2.drawContours(mask1, [c], -1, 5, -1)

            output = cv2.bitwise_and(frame, frame, mask=mask1)
            return output

	    ###################################  D3 ends #########################################

        if (det == 'D4'):

            # load the shapes image, convert it to grayscale, and edge edges in
            # the image
            image = image_array 
	    if col != 'C2':
            	gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
            	edged = cv2.Canny(gray, 50, 100)
	    else:
		edged = cv2.Canny(image, 50, 100)
            

            # find contours in the image and initialize the mask that will be
            # used to remove the bad contours
            (_, cnts, _) = cv2.findContours(
                edged.copy(), cv2.RETR_LIST, cv2.CHAIN_APPROX_SIMPLE)
            mask = np.ones(image.shape[:2], dtype="uint8") * 255

            # loop over the contours
            for c in cnts:
                # if the contour is bad, draw it on the mask
                if is_contour_bad(c):
                    cv2.drawContours(mask, [c], -1, 0, -1)

            # remove the contours from the image
            image = cv2.bitwise_and(image, image, mask=mask)
            return image




############################grpc funcs################################################


######global, configurable#################
initialSize = 800000 #800K
frameRate = 5 #5fps

######global, non configurable############
targetLat = 0
targetAcc = 0
targetDataset = {}
jaadSimpleHT = {}
jaadMediumHT = {}
jaadComplexHT = {}
dukeSimpleHT = {}
dukeMediumHT = {}
dukeComplexHT = {}
sumLatError = 0
firstFrame = ""
firstFramegray = np.zeros(5)
FknobSettings = []
prev_frame = []





class LatencyControllerServicer(controller_api_pb2_grpc.LatencyControllerServicer):
	def SetTarget(self, request, context):
		global targetLat 
		global targetDataset
		global targetAcc  

		targetLat = float(request.target_lat)
		targetAcc, targetDataset = chooseDataset(request.target_acc)

		status = controller_api_pb2.Status(status=True)
		return status
		
		




	def Control(self, request_iterator, context):
		global prev_frame
		success, encoded_image = cv2.imencode('.png', firstFrame)
		image_bytes = encoded_image.tobytes()
		res_image = Image.open(io.BytesIO(image_bytes))
        	res_image = res_image.resize((1920, 1080))
		res_image = np.array(res_image, dtype=np.uint8)
		prev_frame = res_image
		currentLat = 0
		currLatAvg = 0
		imCount = 0
		#prevImSize = initialSize
		prevImSize = findInitialSize(targetLat)
		knob = ["'R2'", "'C1'", "'K1'", "'D1'", "'F1'"]
		acheivedAcc = "0.4" #max accuracy
		for im in request_iterator:			
			imCount = imCount+1
			imRecTimeAndCurrLat = im.current_lat.split('and')
			currentLat += float(imRecTimeAndCurrLat[1])
			if imCount == frameRate:
				currLatAvg = currentLat/frameRate
						
				sizeDelta = findSizeDelta(currLatAvg, targetLat)
				if currLatAvg - targetLat > 0:
					if prevImSize != targetDataset.items()[0][0]:
						newImSize = prevImSize - sizeDelta
					else:
						newImSize = prevImSize
				else:
					if prevImSize != targetDataset.items()[len(targetDataset)-1][0]:
						newImSize = prevImSize + sizeDelta
					else:
						newImSize = prevImSize
				newImSize, knob, acheivedAcc = findKnobs(newImSize)
				acheivedAcc = str(acheivedAcc)
				prevImSize = newImSize

				imCount = 0
				currentLat = 0
	
			modImBytes = modifyImage(knob, im.image)
			if len(modImBytes) == 0:
				continue
			
			response = controller_api_pb2.CustomImage()
			response.image = modImBytes
			response.acheived_acc = imRecTimeAndCurrLat[0]+"and"+acheivedAcc


            		yield response
			
			

		







#starts grpc server
def serve():
	server = grpc.server(futures.ThreadPoolExecutor(max_workers=2))
	controller_api_pb2_grpc.add_LatencyControllerServicer_to_server(LatencyControllerServicer(), server)
	server.add_insecure_port('[::]:9002')
    	server.start()
	print("controller started")
	try:
        	while True:
            		time.sleep(5)
	except KeyboardInterrupt:
        	server.stop(0)




if __name__ == '__main__':
    logging.basicConfig()
loadData()
serve()
