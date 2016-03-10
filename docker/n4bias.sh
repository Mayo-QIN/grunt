#!/bin/bash
#
dim=3 # image dimensionality for now not an argument
AP="/tmp/bin/" # /home/yourself/code/ANTS/bin/bin/  # path to ANTs binaries
ITK_GLOBAL_DEFAULT_NUMBER_OF_THREADS=24  # controls multi-threading
export ITK_GLOBAL_DEFAULT_NUMBER_OF_THREADS



while getopts ":f:o:" opt; do
    case $opt in
        f)
            f=$OPTARG
            ;;
        o)  
            out=$OPTARG
            ;; 
    esac
done 

# Now that we are done parsing options from the command line, shift
# the parsed parameters out of the command line arguments
shift $((OPTIND-1))

if [[ ! -s $f ]] ; then echo no fixed $f ; exit; fi
reg=${AP}N4BiasFieldCorrection           # path to antsRegistration
$reg --bspline-fitting [ 300, 5 ] -d 3 --input-image $f --convergence [ 50x50x30x20, 1e-06 ] --output $out --shrink-factor 3