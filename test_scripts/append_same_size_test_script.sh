cd ../storage

echo -------------TestAppendSameSizeImage-START------------------------

echo FOR 2.1M, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/2.1M/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/2.1M/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/2.1M/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 1.5M, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.5M/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.5M/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.5M/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 1.2M, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.2M/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.2M/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.2M/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 1.0M, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.0M/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.0M/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1.0M/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 800K, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/800K/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/800K/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/800K/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 500K, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/500K/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/500K/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/500K/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 200K, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/200K/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/200K/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/200K/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 100K, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/100K/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/100K/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/100K/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 50K, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/50K/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/50K/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/50K/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"



echo FOR 10K, 30fps, 24fps and 15fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/10K/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/10K/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/10K/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendOverWrittenLog"

echo -------------TestAppendSameSizeImage-DONE-------------------------




echo -------------TestAppendVariableSizeImage-START-------------------------

echo FOR image size between 2.1M to 1K at 30fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendVariableSizeImage"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendVariableSizeImage"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 go test -v -run="TestAppendVarSizeOverWrittenLog"

echo -------------TestAppendVariableSizeImage-DONE-------------------------



echo -------------TestReadAppended-START-------------------------

echo FOR image size between 2.1M to 1K at 30fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=1000 LS=10 SS=100 FILL=NO go test -v -run="TestReadAppended"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=500 LS=10 SS=100 FILL= NO go test -v -run="TestReadAppended"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 FILL=O go test -v -run="TestReadAppended"
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 FILL=O1 go test -v -run="TestReadAppended"
#FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 FILL=O2 go test -v -run="TestReadAppended"

echo -------------TestReadAppended-DONE-------------------------



echo -------------TestReadMultipleSameTstartAndTEnd-START-------------------------

echo FOR image size between 2.1M to 1K at 30fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=1000 LS=10 SS=100 FILL=NO NR=5 go test -v -run="TestReadMultipleSameTstartAndTEnd"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=500 LS=10 SS=100 FILL=NO NR=10 go test -v -run="TestReadMultipleSameTstartAndTEnd"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 FILL=O NR=10 go test -v -run="TestReadMultipleSameTstartAndTEnd"

echo -------------TestReadMultipleSameTstartAndTEnd-DONE-------------------------





echo -------------TestReadMultipleVariableTstartAndTEnd-START-------------------------

echo FOR image size between 2.1M to 1K at 30fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=1000 LS=10 SS=100 FILL=NO NR=5 go test -v -run="TestReadMultipleVariableTstartAndTEnd"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=500 LS=10 SS=100 FILL=NO NR=5 go test -v -run="TestReadMultipleVariableTstartAndTEnd"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 FILL=O NR=5 go test -v -run="TestReadMultipleVariableTstartAndTEnd"

echo -------------TestReadMultipleVariableTstartAndTEnd-DONE-------------------------





echo -------------TestReadConcurrent-START-------------------------

echo FOR image size between 2.1M to 1K at 30fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=1000 LS=10 SS=100 FILL=NO go test -v -run="TestReadConcurrent"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=500 LS=10 SS=100 FILL= NO go test -v -run="TestReadConcurrent"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 FILL=O go test -v -run="TestReadConcurrent"

echo -------------TestReadConcurrent-DONE-------------------------





echo -------------TestReadMultipleConcurrentSameTstartAndTEnd-START-------------------------

echo FOR image size between 2.1M to 1K at 30fps
echo TestFullyWritten
echo logsize - 1000, number of images - 1000
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=1000 LS=10 SS=100 FILL=NO NR=5 go test -v -run="TestReadMultipleConcurrentSameTstartAndTEnd"
echo TestPartiallyWritten
echo logsize - 1000, number of images - 500
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=500 LS=10 SS=100 FILL=NO NR=10 go test -v -run="TestReadMultipleConcurrentSameTstartAndTEnd"
echo TestOverWritten
echo logsize - 300, number of images - 545
FP=/home/research/goworkspace/src/github.com/arun-ravindran/test_images/1000_images/ NO_IM=545 LS=10 SS=30 FILL=O NR=10 go test -v -run="TestReadMultipleConcurrentSameTstartAndTEnd"

echo -------------TestReadMultipleConcurrentSameTstartAndTEnd-DONE-------------------------














