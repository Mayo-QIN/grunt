package main

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestReadXML(t *testing.T) {
	var executable Executable
	parser := xml.NewDecoder(strings.NewReader(ADXML))
	err := parser.Decode(&executable)
	if err != nil {
		t.Errorf("failed to parse %v", err.Error())
	}
	if executable.Title != "Anisotropic Diffusion" {
		t.Errorf("failed to parse title got '%v'", executable.Title)
	}
	if len(executable.ParameterGroups) != 2 {
		t.Errorf("incorrect number of parameter groups found '%v'", len(executable.ParameterGroups))
	}
}

func TestParseService(t *testing.T) {
	for _, tt := range serviceTests {
		s, err := CreateServiceFromXML(tt.executable, tt.xml)
		if err != nil {
			t.Errorf("failed to parse %v", err.Error())
		}
		if len(s.Parameters) != tt.parameter_count {
			t.Errorf("%v unexpected commandline length expected: %v actual: %v -- %v", tt.executable, tt.parameter_count, len(s.Parameters), s.Parameters)
		}
		if len(s.OutputFiles) != tt.output_file_count {
			t.Errorf("%v unexpected output file length expected: %v actual: %v -- %v", tt.executable, tt.output_file_count, len(s.OutputFiles), s.OutputFiles)
		}
		if len(s.InputFiles) != tt.input_file_count {
			t.Errorf("%v unexpected input file length expected: %v actual: %v -- %v", tt.executable, tt.input_file_count, len(s.InputFiles), s.InputFiles)
		}
	}
}

var serviceTests = []struct {
	title             string
	executable        string
	parameter_count   int
	input_file_count  int
	output_file_count int
	xml               string
}{
	{"Anisotropic Diffusion", "AnisotropicDiffusion", 3, 1, 1, ADXML},
	{"Gradient Anisotropic Diffusion", "GradientAnisotropicDiffusion", 4, 1, 1, GADXML},
	{"Tour", "ExecutionModelTour", 10, 2, 4, TourXML},
	// NB: this is an error, the segimage2itkimage has a bug!  The OutputDirName should be an "output" channel
	{"Seg Image 2 ITK Image", "segimage2itkimage", 2, 2, 0, segimage2itkimageXML},
}

// These XML strings are from the Slicer CLI modules and website examples.
var ADXML = `<?xml version="1.0" encoding="utf-8"?>
<executable>
  <category>filtering</category>
  <title>Anisotropic Diffusion</title>
  <description>
  Runs anisotropic diffusion on a volume
  </description>
  <version>1.0</version>
  <documentation-url></documentation-url>
  <license></license>
  <contributor>Bill Lorensen</contributor>

  <parameters>
    <label>
    Anisotropic Diffusion Parameters
    </label>
    <description>
    Parameters for the anisotropic
    diffusion algorithm
    </description>

    <double>
      <name>conductance</name>
      <longflag>conductance</longflag>
      <description>Conductance</description>
      <label>Conductance</label>
      <default>1</default>
      <constraints>
        <minimum>0</minimum>
        <maximum>10</maximum>
        <step>.01</step>
      </constraints>
    </double>

    <double>
      <name>timeStep</name>
      <longflag>timeStep</longflag>
      <description>Time Step</description>
      <label>Time Step</label>
      <default>0.0625</default>
      <constraints>
        <minimum>.001</minimum>
        <maximum>1</maximum>
        <step>.001</step>
      </constraints>
    </double>

    <integer>
      <name>numberOfIterations</name>
      <longflag>iterations</longflag>
      <description>Number of iterations</description>
      <label>Iterations</label>
      <default>1</default>
      <constraints>
        <minimum>1</minimum>
        <maximum>30</maximum>
        <step>1</step>
      </constraints>
    </integer>

  </parameters>

  <parameters>
    <label>IO</label>
    <description>Input/output parameters</description>
    <image>
      <name>inputVolume</name>
      <label>Input Volume</label>
      <channel>input</channel>
      <index>0</index>
      <description>Input volume to be filtered</description>
    </image>
    <image>
      <name>outputVolume</name>
      <label>Output Volume</label>
      <channel>output</channel>
      <index>1</index>
      <description>Output filtered</description>
    </image>
  </parameters>

</executable>`

var GADXML = `<?xml version="1.0" encoding="utf-8"?>
<executable>
  <category>Filtering.Denoising</category>
  <index>1</index>
  <title>Gradient Anisotropic Diffusion</title>
  <description><![CDATA[Runs gradient anisotropic diffusion on a volume.

Anisotropic diffusion methods reduce noise (or unwanted detail) in images while preserving specific image features, like edges.  For many applications, there is an assumption that light-dark transitions (edges) are interesting.  Standard isotropic diffusion methods move and blur light-dark boundaries.  Anisotropic diffusion methods are formulated to specifically preserve edges. The conductance term for this implementation is a function of the gradient magnitude of the image at each point, reducing the strength of diffusion at edges. The numerical implementation of this equation is similar to that described in the Perona-Malik paper, but uses a more robust technique for gradient magnitude estimation and has been generalized to N-dimensions.]]></description>
  <version>0.1.0.$Revision: 24424 $(alpha)</version>
  <documentation-url>http://wiki.slicer.org/slicerWiki/index.php/Documentation/Nightly/Modules/GradientAnisotropicDiffusion</documentation-url>
  <license/>
  <contributor>Bill Lorensen (GE)</contributor>
  <acknowledgements><![CDATA[This command module was derived from Insight/Examples (copyright) Insight Software Consortium]]></acknowledgements>
  <parameters>
    <label>Anisotropic Diffusion Parameters</label>
    <description><![CDATA[Parameters for the anisotropic diffusion algorithm]]></description>
    <double>
      <name>conductance</name>
      <longflag>--conductance</longflag>
      <description><![CDATA[Conductance controls the sensitivity of the conductance term. As a general rule, the lower the value, the more strongly the filter preserves edges. A high value will cause diffusion (smoothing) across edges. Note that the number of iterations controls how much smoothing is done within regions bounded by edges.]]></description>
      <label>Conductance</label>
      <default>1</default>
      <constraints>
        <minimum>0</minimum>
        <maximum>10</maximum>
        <step>.01</step>
      </constraints>
    </double>
    <integer>
      <name>numberOfIterations</name>
      <longflag>--iterations</longflag>
      <description><![CDATA[The more iterations, the more smoothing. Each iteration takes the same amount of time. If it takes 10 seconds for one iteration, then it will take 100 seconds for 10 iterations. Note that the conductance controls how much each iteration smooths across edges.]]></description>
      <label>Iterations</label>
      <default>5</default>
      <constraints>
        <minimum>1</minimum>
        <maximum>30</maximum>
        <step>1</step>
      </constraints>
    </integer>
    <double>
      <name>timeStep</name>
      <longflag>--timeStep</longflag>
      <description><![CDATA[The time step depends on the dimensionality of the image. In Slicer the images are 3D and the default (.0625) time step will provide a stable solution.]]></description>
      <label>Time Step</label>
      <default>0.0625</default>
      <constraints>
        <minimum>.001</minimum>
        <maximum>.0625</maximum>
        <step>.001</step>
      </constraints>
    </double>
  </parameters>
  <parameters>
    <label>IO</label>
    <description><![CDATA[Input/output parameters]]></description>
    <image>
      <name>inputVolume</name>
      <label>Input Volume</label>
      <channel>input</channel>
      <index>0</index>
      <description><![CDATA[Input volume to be filtered]]></description>
    </image>
    <image>
      <name>outputVolume</name>
      <label>Output Volume</label>
      <channel>output</channel>
      <index>1</index>
      <description><![CDATA[Output filtered]]></description>
    </image>
  </parameters>
  <parameters advanced = "true">
    <label>Advanced</label>
    <description><![CDATA[Advanced parameters for the anisotropic diffusion algorithm]]></description>
    <boolean>
      <name>useImageSpacing</name>
      <longflag>--useImageSpacing</longflag>
      <description>![CDATA[Take into account image spacing in the computation.  It is advisable to turn this option on, especially when the pixel size is different in different dimensions. However, to produce results consistent with Slicer4.2 and earlier, this option should be turned off.]]</description>
      <label>Use image spacing</label>
      <default>true</default>
    </boolean>
  </parameters>
</executable>
`
var TourXML = `<?xml version="1.0" encoding="utf-8"?>
<executable>
  <category>Developer Tools</category>
  <title>Execution Model Tour</title>
  <description><![CDATA[Shows one of each type of parameter.]]></description>
  <version>0.1.0.$Revision: 24459 $(alpha)</version>
  <documentation-url>http://wiki.slicer.org/slicerWiki/index.php/Documentation/Nightly/Modules/ExecutionModelTour</documentation-url>
  <license/>
  <contributor>Daniel Blezek (GE), Bill Lorensen (GE)</contributor>
  <acknowledgements><![CDATA[This work is part of the National Alliance for Medical Image Computing (NAMIC), funded by the National Institutes of Health through the NIH Roadmap for Medical Research, Grant U54 EB005149.]]></acknowledgements>
  <parameters>
    <label>Scalar Parameters</label>
    <description><![CDATA[Variations on scalar parameters]]></description>
    <integer>
      <name>integerVariable</name>
      <flag>-i</flag>
      <longflag>--integer</longflag>
      <description><![CDATA[An integer without constraints]]></description>
      <label>Integer Parameter</label>
      <default>30</default>
    </integer>
    <double>
      <name>doubleVariable</name>
      <flag>-d</flag>
      <longflag>--double</longflag>
      <description><![CDATA[A double with constraints]]></description>
      <label>Double Parameter</label>
      <default>30</default>
      <constraints>
        <minimum>0</minimum>
        <maximum>1.e3</maximum>
        <step>10</step>
      </constraints>
    </double>
  </parameters>
  <parameters>
    <label>Vector Parameters</label>
    <description><![CDATA[Variations on vector parameters]]></description>
    <float-vector>
      <name>floatVector</name>
      <flag>f</flag>
      <description><![CDATA[A vector of floats]]></description>
      <label>Float Vector Parameter</label>
      <default>1.3,2,-14</default>
    </float-vector>
    <string-vector>
      <name>stringVector</name>
      <longflag>string_vector</longflag>
      <description><![CDATA[A vector of strings]]></description>
      <label>String Vector Parameter</label>
      <default>foo,bar,foobar</default>
    </string-vector>
  </parameters>
  <parameters>
    <label>Enumeration Parameters</label>
    <description><![CDATA[Variations on enumeration parameters]]></description>
    <string-enumeration>
      <name>stringChoice</name>
      <flag>e</flag>
      <longflag>enumeration</longflag>
      <description><![CDATA[An enumeration of strings]]></description>
      <label>String Enumeration Parameter</label>
      <default>Bill</default>
      <element>Ron</element>
      <element>Eric</element>
      <element>Bill</element>
      <element>Ross</element>
      <element>Steve</element>
      <element>Will</element>
    </string-enumeration>
  </parameters>
  <parameters>
    <label>Boolean Parameters</label>
    <description><![CDATA[Variations on boolean parameters]]></description>
    <boolean>
      <name>boolean1</name>
      <longflag>boolean1</longflag>
      <description><![CDATA[A boolean default true]]></description>
      <label>Boolean Default true</label>
      <default>true</default>
    </boolean>
    <boolean>
      <name>boolean2</name>
      <longflag>boolean2</longflag>
      <description><![CDATA[A boolean default false]]></description>
      <label>Boolean Default false</label>
      <default>false</default>
    </boolean>
    <boolean>
      <name>boolean3</name>
      <longflag>boolean3</longflag>
      <description><![CDATA[A boolean with no default, should be defaulting to false]]></description>
      <label>Boolean No Default</label>
    </boolean>
  </parameters>
  <parameters>
    <label>File, Directory and Image Parameters</label>
    <description><![CDATA[Parameters that describe files and direcories.]]></description>
    <file fileExtensions=".png,.jpg,.jpeg,.bmp,.tif,.tiff,.gipl,.dcm,.dicom,.nhdr,.nrrd,.mhd,.mha,.mask,.hdr,.nii,.nii.gz,.hdr.gz,.pic,.lsm,.spr,.vtk,.vtkp,.vtki,.stl,.csv,.txt,.xml,.html">
      <longflag>file1</longflag>
      <description><![CDATA[An input file]]></description>
      <label>Input file</label>
      <channel>input</channel>
    </file>
    <file fileExtensions=".png,.jpg,.jpeg,.bmp,.tif,.tiff,.gipl,.dcm,.dicom,.nhdr,.nrrd,.mhd,.mha,.mask,.hdr,.nii,.nii.gz,.hdr.gz,.pic,.lsm,.spr,.vtk,.vtkp,.vtki,.stl,.csv,.txt,.xml,.html" multiple="true">
      <longflag>files</longflag>
      <description><![CDATA[Multiple input files]]></description>
      <label>Input Files</label>
      <channel>input</channel>
    </file>
    <directory>
      <longflag>directory1</longflag>
      <description><![CDATA[An input directory. If no default is specified, the current directory is used,]]></description>
      <label>Input directory</label>
      <channel>input</channel>
    </directory>
    <image>
      <longflag>image1</longflag>
      <description><![CDATA[An input image]]></description>
      <label>Input image</label>
      <channel>input</channel>
    </image>
    <image>
      <longflag>image2</longflag>
      <description><![CDATA[An output image]]></description>
      <label>Output image</label>
      <channel>output</channel>
    </image>
    <transform type="linear">
      <longflag>transform1</longflag>
      <description><![CDATA[An input transform]]></description>
      <label>Input transform</label>
      <channel>input</channel>
    </transform>
    <transform type="linear">
      <longflag>transform2</longflag>
      <description><![CDATA[An output transform]]></description>
      <label>Output transform</label>
      <channel>output</channel>
    </transform>
    <point multiple="true" coordinateSystem="ras">
      <name>seed</name>
      <label>Seeds</label>
      <longflag>--seed</longflag>
      <description><![CDATA[Lists of points in the CLI correspond to slicer fiducial lists]]></description>
      <default>0,0,0</default>
    </point>
    <pointfile multiple="true" fileExtensions=".fcsv" coordinateSystem="lps">
      <name>seedsFile</name>
      <description><![CDATA[Test file of input fiducials, compared to seeds]]></description>
      <label>Seeds file</label>
      <longflag>seedsFile</longflag>
      <channel>input</channel>
    </pointfile>
  </parameters>
  <parameters>
    <label>Index Parameters</label>
    <description><![CDATA[Variations on parameters that use index rather than flags.]]></description>
    <image>
      <name>arg0</name>
      <channel>input</channel>
      <index>0</index>
      <description><![CDATA[First index argument is an image]]></description>
      <label>First index argument</label>
    </image>
    <image>
      <name>arg1</name>
      <channel>output</channel>
      <index>1</index>
      <description><![CDATA[Second index argument is an image]]></description>
      <label>Second index argument</label>
    </image>
  </parameters>
  <parameters>
    <label>Regions of interest</label>
    <region multiple="true">
      <name>regions</name>
      <label>Region list</label>
      <longflag>region</longflag>
      <description><![CDATA[List of regions to process]]></description>
    </region>
  </parameters>
  <parameters>
    <label>Measurements</label>
    <measurement>
      <name>inputFA</name>
      <channel>input</channel>
      <label>Input FA measurements</label>
      <longflag>inputFA</longflag>
      <description><![CDATA[Array of FA values to process]]></description>
    </measurement>
    <measurement>
      <name>outputFA</name>
      <channel>output</channel>
      <label>Output FA measurements</label>
      <longflag>outputFA</longflag>
      <description><![CDATA[Array of processed (output) FA values]]></description>
    </measurement>
  </parameters>
  <parameters>
    <label>Simple return types</label>
    <integer>
      <name>anintegerreturn</name>
      <label>An integer return value</label>
      <channel>output</channel>
      <default>5</default>
      <description><![CDATA[An example of an integer return type]]></description>
    </integer>
    <boolean>
      <name>abooleanreturn</name>
      <label>A boolean return value</label>
      <channel>output</channel>
      <default>false</default>
      <description><![CDATA[An example of a boolean return type]]></description>
    </boolean>
    <float>
      <name>afloatreturn</name>
      <label>A floating point return value</label>
      <channel>output</channel>
      <default>7.0</default>
      <description><![CDATA[An example of a float return type]]></description>
    </float>
    <double>
      <name>adoublereturn</name>
      <label>A double point return value</label>
      <channel>output</channel>
      <default>14.0</default>
      <description><![CDATA[An example of a double return type]]></description>
    </double>
    <string>
      <name>astringreturn</name>
      <label>A string point return value</label>
      <channel>output</channel>
      <default>Hello</default>
      <description><![CDATA[An example of a string return type]]></description>
    </string>
    <integer-vector>
      <name>anintegervectorreturn</name>
      <label>An integer vector return value</label>
      <channel>output</channel>
      <default>1,2,3</default>
      <description><![CDATA[An example of an integer vector return type]]></description>
    </integer-vector>
    <string-enumeration>
      <name>astringchoicereturn</name>
      <channel>output</channel>
      <description><![CDATA[An enumeration of strings as a return type]]></description>
      <label>A string enumeration return value</label>
      <default>Bill</default>
      <element>Ron</element>
      <element>Eric</element>
      <element>Bill</element>
      <element>Ross</element>
      <element>Steve</element>
      <element>Will</element>
    </string-enumeration>
  </parameters>
  <parameters>
    <label>File return types</label>
    <pointfile fileExtensions=".fcsv" coordinateSystem="lps">
      <name>seedsOutFile</name>
      <label>Output Fiducials File</label>
      <description><![CDATA[Output file to read back in, compare to seeds with flipped settings on first fiducial]]></description>
      <longflag>seedsOutFile</longflag>
      <channel>output</channel>
    </pointfile>
  </parameters>
</executable>

`
var segimage2itkimageXML = `<?xml version="1.0" encoding="utf-8"?>
<executable>
  <category>Informatics</category>
  <title>Convert DICOM SEG into ITK image</title>
  <description>This tool can be used to convert DICOM Segmentation into volumetric segmentations stored as labeled pixels using research format, such as NRRD or NIfTI, and meta information stored in the JSON file format.</description>
  <version>1.0</version>
  <documentation-url>https://github.com/QIICR/dcmqi</documentation-url>
  <license></license>
  <contributor>Andrey Fedorov(BWH), Christian Herz(BWH)</contributor>
  <acknowledgements>This work is supported in part the National Institutes of Health, National Cancer Institute, Informatics Technology for Cancer Research (ITCR) program, grant Quantitative Image Informatics for Cancer Research (QIICR) (U24 CA180918, PIs Kikinis and Fedorov).</acknowledgements>

  <parameters>
    <file>
      <name>inputSEGFileName</name>
      <label>SEG file name</label>
      <channel>input</channel>
      <longflag>inputDICOM</longflag>
      <description>File name of the input DICOM Segmentation image object.</description>
    </file>

    <file>
      <name>outputDirName</name>
      <label>Output directory name</label>
      <channel>input</channel>
      <longflag>outputDirectory</longflag>
      <description>Directory to store individual segments saved using the output format specified files. When specified, file names will contain prefix, followed by the segment number.</description>
    </file>

    <string>
      <name>prefix</name>
      <label>Output prefix</label>
      <flag>p</flag>
      <longflag>prefix</longflag>
      <description>Prefix for output file.</description>
      <default></default>
    </string>

    <string-enumeration>
      <name>outputType</name>
      <flag>t</flag>
      <longflag>outputType</longflag>
      <description>Output file format for the resulting image data.</description>
      <label>Output type</label>
      <default>nrrd</default>
      <element>nrrd</element>
      <element>mhd</element>
      <element>mha</element>
      <element>nii</element>
      <element>nifti</element>
      <element>hdr</element>
      <element>img</element>
    </string-enumeration>

  </parameters>

</executable>
`
