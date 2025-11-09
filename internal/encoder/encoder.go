package encoder

import (
	"errors"
	"fmt"

	"github.com/csnewman/ffmpeg-go"
)

// Encoder handles MP3 encoding from audio input files
type Encoder struct {
	inputPath  string
	outputPath string
	stereo     bool

	ifmtCtx *ffmpeg.AVFormatContext
	ofmtCtx *ffmpeg.AVFormatContext

	decCtx *ffmpeg.AVCodecContext
	encCtx *ffmpeg.AVCodecContext

	decFrame *ffmpeg.AVFrame
	encPkt   *ffmpeg.AVPacket

	filterGraph   *ffmpeg.AVFilterGraph
	bufferSrcCtx  *ffmpeg.AVFilterContext
	bufferSinkCtx *ffmpeg.AVFilterContext
	filteredFrame *ffmpeg.AVFrame

	streamIndex  int
	samplesRead  int64
	totalSamples int64
	nextPts      int64 // Track PTS for output frames
}

// Config holds encoder configuration
type Config struct {
	InputPath  string
	OutputPath string
	Stereo     bool // true = 192kbps stereo, false = 112kbps mono
}

// New creates a new encoder instance
func New(cfg Config) (*Encoder, error) {
	if cfg.InputPath == "" {
		return nil, fmt.Errorf("input path is required")
	}
	if cfg.OutputPath == "" {
		return nil, fmt.Errorf("output path is required")
	}

	return &Encoder{
		inputPath:   cfg.InputPath,
		outputPath:  cfg.OutputPath,
		stereo:      cfg.Stereo,
		streamIndex: -1,
	}, nil
}

// Initialize opens input and output files, sets up decoder and encoder
func (e *Encoder) Initialize() error {
	// Suppress FFmpeg logs during normal operation
	ffmpeg.AVLogSetLevel(ffmpeg.AVLogError)

	if err := e.openInput(); err != nil {
		return fmt.Errorf("failed to open input: %w", err)
	}

	if err := e.openOutput(); err != nil {
		e.Close()
		return fmt.Errorf("failed to open output: %w", err)
	}

	// Allocate frames and packets
	e.decFrame = ffmpeg.AVFrameAlloc()
	e.filteredFrame = ffmpeg.AVFrameAlloc()
	e.encPkt = ffmpeg.AVPacketAlloc()

	// Initialize audio filter graph
	if err := e.initFilter(); err != nil {
		return fmt.Errorf("failed to initialize filter: %w", err)
	}

	return nil
}

// openInput opens and analyzes the input audio file
func (e *Encoder) openInput() error {
	// Open input file
	urlPtr := ffmpeg.ToCStr(e.inputPath)
	defer urlPtr.Free()

	if _, err := ffmpeg.AVFormatOpenInput(&e.ifmtCtx, urlPtr, nil, nil); err != nil {
		return fmt.Errorf("cannot open input file: %w", err)
	}

	// Find stream information
	if _, err := ffmpeg.AVFormatFindStreamInfo(e.ifmtCtx, nil); err != nil {
		return fmt.Errorf("cannot find stream information: %w", err)
	}

	// Find best audio stream
	streamIdx, err := ffmpeg.AVFindBestStream(e.ifmtCtx, ffmpeg.AVMediaTypeAudio, -1, -1, nil, 0)
	if err != nil {
		return fmt.Errorf("cannot find audio stream: %w", err)
	}
	e.streamIndex = streamIdx

	stream := e.ifmtCtx.Streams().Get(uintptr(e.streamIndex))
	codecPar := stream.Codecpar()

	// Find decoder
	decoder := ffmpeg.AVCodecFindDecoder(codecPar.CodecId())
	if decoder == nil {
		return fmt.Errorf("decoder not found for codec %d", codecPar.CodecId())
	}

	// Allocate decoder context
	e.decCtx = ffmpeg.AVCodecAllocContext3(decoder)
	if e.decCtx == nil {
		return fmt.Errorf("failed to allocate decoder context")
	}

	// Copy codec parameters to decoder context
	if _, err := ffmpeg.AVCodecParametersToContext(e.decCtx, codecPar); err != nil {
		return fmt.Errorf("failed to copy codec parameters: %w", err)
	}

	// Open decoder
	if _, err := ffmpeg.AVCodecOpen2(e.decCtx, decoder, nil); err != nil {
		return fmt.Errorf("failed to open decoder: %w", err)
	}

	// Calculate total samples for progress tracking
	duration := stream.Duration()
	timeBase := stream.TimeBase()
	if duration > 0 {
		durationSec := float64(duration) * float64(timeBase.Num()) / float64(timeBase.Den())
		e.totalSamples = int64(durationSec * float64(e.decCtx.SampleRate()))
	}

	return nil
}

// openOutput creates the output MP3 file and sets up the encoder
func (e *Encoder) openOutput() error {
	// Allocate output format context
	namePtr := ffmpeg.ToCStr(e.outputPath)
	defer namePtr.Free()

	if _, err := ffmpeg.AVFormatAllocOutputContext2(&e.ofmtCtx, nil, nil, namePtr); err != nil {
		return fmt.Errorf("failed to create output context: %w", err)
	}

	// Find MP3 encoder
	encoder := ffmpeg.AVCodecFindEncoder(ffmpeg.AVCodecIdMp3)
	if encoder == nil {
		return fmt.Errorf("MP3 encoder not found")
	}

	// Create output stream
	outStream := ffmpeg.AVFormatNewStream(e.ofmtCtx, encoder)
	if outStream == nil {
		return fmt.Errorf("failed to create output stream")
	}

	// Allocate encoder context
	e.encCtx = ffmpeg.AVCodecAllocContext3(encoder)
	if e.encCtx == nil {
		return fmt.Errorf("failed to allocate encoder context")
	}

	// Configure encoder for podcast MP3
	if e.stereo {
		// Stereo mode: 192kbps
		e.encCtx.SetBitRate(192000)
		e.encCtx.SetChannels(2)
		e.encCtx.SetChannelLayout(ffmpeg.AVChLayoutStereo)
	} else {
		// Mono mode: 112kbps
		e.encCtx.SetBitRate(112000)
		e.encCtx.SetChannels(1)
		e.encCtx.SetChannelLayout(ffmpeg.AVChLayoutMono)
	}

	e.encCtx.SetSampleRate(44100)
	e.encCtx.SetSampleFmt(ffmpeg.AVSampleFmtS16P) // Signed 16-bit planar

	// Create time base
	tb := &ffmpeg.AVRational{}
	tb.SetNum(1)
	tb.SetDen(e.encCtx.SampleRate())
	e.encCtx.SetTimeBase(tb)

	// Set LAME-specific options via AVDictionary
	var opts *ffmpeg.AVDictionary

	keyComp := ffmpeg.ToCStr("compression_level")
	defer keyComp.Free()
	valComp := ffmpeg.ToCStr("3") // LAME quality preset -q 3
	defer valComp.Free()
	if _, err := ffmpeg.AVDictSet(&opts, keyComp, valComp, 0); err != nil {
		return fmt.Errorf("failed to set compression_level: %w", err)
	}

	keyCutoff := ffmpeg.ToCStr("cutoff")
	defer keyCutoff.Free()
	valCutoff := ffmpeg.ToCStr("20500") // 20.5kHz lowpass cutoff frequency
	defer valCutoff.Free()
	if _, err := ffmpeg.AVDictSet(&opts, keyCutoff, valCutoff, 0); err != nil {
		ffmpeg.AVDictFree(&opts)
		return fmt.Errorf("failed to set cutoff: %w", err)
	}

	// Open encoder
	if _, err := ffmpeg.AVCodecOpen2(e.encCtx, encoder, &opts); err != nil {
		ffmpeg.AVDictFree(&opts)
		return fmt.Errorf("failed to open encoder: %w", err)
	}
	ffmpeg.AVDictFree(&opts)

	// Copy encoder parameters to output stream
	if _, err := ffmpeg.AVCodecParametersFromContext(outStream.Codecpar(), e.encCtx); err != nil {
		return fmt.Errorf("failed to copy encoder parameters: %w", err)
	}

	outStream.SetTimeBase(e.encCtx.TimeBase())

	// Open output file
	if e.ofmtCtx.Oformat().Flags()&ffmpeg.AVFmtNofile == 0 {
		var pb *ffmpeg.AVIOContext
		if _, err := ffmpeg.AVIOOpen(&pb, e.ofmtCtx.Url(), ffmpeg.AVIOFlagWrite); err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		e.ofmtCtx.SetPb(pb)
	}

	// Write header
	if _, err := ffmpeg.AVFormatWriteHeader(e.ofmtCtx, nil); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// initFilter sets up audio filter graph for resampling and frame buffering
func (e *Encoder) initFilter() error {
	e.filterGraph = ffmpeg.AVFilterGraphAlloc()
	if e.filterGraph == nil {
		return fmt.Errorf("failed to allocate filter graph")
	}

	// Get abuffer and abuffersink filters
	bufferSrc := ffmpeg.AVFilterGetByName(ffmpeg.GlobalCStr("abuffer"))
	bufferSink := ffmpeg.AVFilterGetByName(ffmpeg.GlobalCStr("abuffersink"))
	if bufferSrc == nil || bufferSink == nil {
		return fmt.Errorf("abuffer or abuffersink filter not found")
	}

	// Get channel layout string
	layoutPtr := ffmpeg.AllocCStr(64)
	defer layoutPtr.Free()
	if _, err := ffmpeg.AVChannelLayoutDescribe(e.decCtx.ChLayout(), layoutPtr, 64); err != nil {
		return fmt.Errorf("failed to describe channel layout: %w", err)
	}

	// Create abuffer source args
	pktTimebase := e.decCtx.PktTimebase()
	args := fmt.Sprintf(
		"time_base=%d/%d:sample_rate=%d:sample_fmt=%s:channel_layout=%s",
		pktTimebase.Num(), pktTimebase.Den(),
		e.decCtx.SampleRate(),
		ffmpeg.AVGetSampleFmtName(e.decCtx.SampleFmt()).String(),
		layoutPtr.String(),
	)

	argsC := ffmpeg.ToCStr(args)
	defer argsC.Free()

	// Create buffer source
	if _, err := ffmpeg.AVFilterGraphCreateFilter(
		&e.bufferSrcCtx,
		bufferSrc,
		ffmpeg.GlobalCStr("in"),
		argsC,
		nil,
		e.filterGraph,
	); err != nil {
		return fmt.Errorf("failed to create buffer source: %w", err)
	}

	// Create buffer sink
	if _, err := ffmpeg.AVFilterGraphCreateFilter(
		&e.bufferSinkCtx,
		bufferSink,
		ffmpeg.GlobalCStr("out"),
		nil,
		nil,
		e.filterGraph,
	); err != nil {
		return fmt.Errorf("failed to create buffer sink: %w", err)
	}

	// Parse filter graph - use aresample for format/rate/channel conversion
	outputs := ffmpeg.AVFilterInoutAlloc()
	inputs := ffmpeg.AVFilterInoutAlloc()
	defer ffmpeg.AVFilterInoutFree(&outputs)
	defer ffmpeg.AVFilterInoutFree(&inputs)

	outputs.SetName(ffmpeg.ToCStr("in"))
	outputs.SetFilterCtx(e.bufferSrcCtx)
	outputs.SetPadIdx(0)
	outputs.SetNext(nil)

	inputs.SetName(ffmpeg.ToCStr("out"))
	inputs.SetFilterCtx(e.bufferSinkCtx)
	inputs.SetPadIdx(0)
	inputs.SetNext(nil)

	// Build filter spec: resample to target format, then convert channels if needed
	var filterSpec string
	if e.stereo {
		// Stereo: aresample to 44100Hz S16P format, keep stereo, set frame size for LAME
		filterSpec = fmt.Sprintf("aresample=%d:async=1,aformat=sample_fmts=s16p:sample_rates=44100:channel_layouts=stereo,asetnsamples=n=1152",
			e.encCtx.SampleRate())
	} else {
		// Mono: aresample to 44100Hz S16P format, downmix to mono, set frame size for LAME
		filterSpec = fmt.Sprintf("aresample=%d:async=1,aformat=sample_fmts=s16p:sample_rates=44100:channel_layouts=mono,asetnsamples=n=1152",
			e.encCtx.SampleRate())
	}

	filterSpecC := ffmpeg.ToCStr(filterSpec)
	defer filterSpecC.Free()

	if _, err := ffmpeg.AVFilterGraphParsePtr(e.filterGraph, filterSpecC, &inputs, &outputs, nil); err != nil {
		return fmt.Errorf("failed to parse filter graph: %w", err)
	}

	// Configure filter graph
	if _, err := ffmpeg.AVFilterGraphConfig(e.filterGraph, nil); err != nil {
		return fmt.Errorf("failed to configure filter graph: %w", err)
	}

	return nil
}

// ProgressCallback is called during encoding with progress updates
type ProgressCallback func(samplesProcessed, totalSamples int64)

// Encode performs the actual encoding with progress callbacks
func (e *Encoder) Encode(progressCb ProgressCallback) error {
	packet := ffmpeg.AVPacketAlloc()
	defer ffmpeg.AVPacketFree(&packet)

	outStream := e.ofmtCtx.Streams().Get(0)

	// Main decoding loop
	for {
		if _, err := ffmpeg.AVReadFrame(e.ifmtCtx, packet); err != nil {
			if errors.Is(err, ffmpeg.AVErrorEOF) {
				break
			}
			return fmt.Errorf("read frame failed: %w", err)
		}

		if packet.StreamIndex() != e.streamIndex {
			ffmpeg.AVPacketUnref(packet)
			continue
		}

		// Decode audio packet
		if _, err := ffmpeg.AVCodecSendPacket(e.decCtx, packet); err != nil {
			ffmpeg.AVPacketUnref(packet)
			return fmt.Errorf("send packet to decoder failed: %w", err)
		}

		ffmpeg.AVPacketUnref(packet)

		// Retrieve decoded frames
		for {
			if _, err := ffmpeg.AVCodecReceiveFrame(e.decCtx, e.decFrame); err != nil {
				if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
					break
				}
				return fmt.Errorf("receive frame from decoder failed: %w", err)
			}

			// Update progress
			e.samplesRead += int64(e.decFrame.NbSamples())
			if progressCb != nil && e.totalSamples > 0 {
				progressCb(e.samplesRead, e.totalSamples)
			}

			// Push frame to filter graph
			if _, err := ffmpeg.AVBuffersrcAddFrameFlags(e.bufferSrcCtx, e.decFrame, ffmpeg.AVBuffersrcFlagKeepRef); err != nil {
				return fmt.Errorf("failed to feed filter graph: %w", err)
			}

			// Pull filtered frames and encode
			for {
				if _, err := ffmpeg.AVBuffersinkGetFrame(e.bufferSinkCtx, e.filteredFrame); err != nil {
					if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
						break
					}
					return fmt.Errorf("failed to get filtered frame: %w", err)
				}

				if err := e.encodeFrame(e.filteredFrame, outStream); err != nil {
					return err
				}

				ffmpeg.AVFrameUnref(e.filteredFrame)
			}

			ffmpeg.AVFrameUnref(e.decFrame)
		}
	}

	// Flush decoder
	if _, err := ffmpeg.AVCodecSendPacket(e.decCtx, nil); err != nil {
		return fmt.Errorf("flush decoder failed: %w", err)
	}

	for {
		if _, err := ffmpeg.AVCodecReceiveFrame(e.decCtx, e.decFrame); err != nil {
			if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
				break
			}
			return fmt.Errorf("flush decoder receive failed: %w", err)
		}

		// Push to filter
		if _, err := ffmpeg.AVBuffersrcAddFrameFlags(e.bufferSrcCtx, e.decFrame, ffmpeg.AVBuffersrcFlagKeepRef); err != nil {
			return fmt.Errorf("failed to feed filter graph: %w", err)
		}

		// Pull filtered frames
		for {
			if _, err := ffmpeg.AVBuffersinkGetFrame(e.bufferSinkCtx, e.filteredFrame); err != nil {
				if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
					break
				}
				return fmt.Errorf("failed to get filtered frame: %w", err)
			}

			if err := e.encodeFrame(e.filteredFrame, outStream); err != nil {
				return err
			}

			ffmpeg.AVFrameUnref(e.filteredFrame)
		}

		ffmpeg.AVFrameUnref(e.decFrame)
	}

	// Flush filter graph
	if _, err := ffmpeg.AVBuffersrcAddFrameFlags(e.bufferSrcCtx, nil, 0); err != nil {
		return fmt.Errorf("failed to flush filter graph: %w", err)
	}

	for {
		if _, err := ffmpeg.AVBuffersinkGetFrame(e.bufferSinkCtx, e.filteredFrame); err != nil {
			if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
				break
			}
			return fmt.Errorf("failed to get filtered frame: %w", err)
		}

		if err := e.encodeFrame(e.filteredFrame, outStream); err != nil {
			return err
		}

		ffmpeg.AVFrameUnref(e.filteredFrame)
	}

	// Flush encoder
	if err := e.flushEncoder(outStream); err != nil {
		return err
	}

	// Write trailer
	if _, err := ffmpeg.AVWriteTrailer(e.ofmtCtx); err != nil {
		return fmt.Errorf("write trailer failed: %w", err)
	}

	return nil
}

// encodeFrame encodes a single audio frame to MP3
func (e *Encoder) encodeFrame(frame *ffmpeg.AVFrame, outStream *ffmpeg.AVStream) error {
	// Set PTS based on our running counter
	frame.SetPts(e.nextPts)

	// Increment PTS by number of samples in this frame
	e.nextPts += int64(frame.NbSamples())

	// Send frame to encoder
	if _, err := ffmpeg.AVCodecSendFrame(e.encCtx, frame); err != nil {
		return fmt.Errorf("send frame to encoder failed: %w", err)
	}

	// Retrieve encoded packets
	for {
		ffmpeg.AVPacketUnref(e.encPkt)

		if _, err := ffmpeg.AVCodecReceivePacket(e.encCtx, e.encPkt); err != nil {
			if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
				break
			}
			return fmt.Errorf("receive packet from encoder failed: %w", err)
		}

		// Set stream index and rescale timestamps
		e.encPkt.SetStreamIndex(0)
		ffmpeg.AVPacketRescaleTs(e.encPkt, e.encCtx.TimeBase(), outStream.TimeBase())

		// Write packet to output
		if _, err := ffmpeg.AVInterleavedWriteFrame(e.ofmtCtx, e.encPkt); err != nil {
			return fmt.Errorf("write frame failed: %w", err)
		}
	}

	return nil
}

// flushEncoder flushes remaining packets from the encoder
func (e *Encoder) flushEncoder(outStream *ffmpeg.AVStream) error {
	if _, err := ffmpeg.AVCodecSendFrame(e.encCtx, nil); err != nil {
		return fmt.Errorf("flush encoder failed: %w", err)
	}

	for {
		ffmpeg.AVPacketUnref(e.encPkt)

		if _, err := ffmpeg.AVCodecReceivePacket(e.encCtx, e.encPkt); err != nil {
			if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
				break
			}
			return fmt.Errorf("flush encoder receive failed: %w", err)
		}

		e.encPkt.SetStreamIndex(0)
		ffmpeg.AVPacketRescaleTs(e.encPkt, e.encCtx.TimeBase(), outStream.TimeBase())

		if _, err := ffmpeg.AVInterleavedWriteFrame(e.ofmtCtx, e.encPkt); err != nil {
			return fmt.Errorf("write frame failed: %w", err)
		}
	}

	return nil
}

// Close releases all resources
func (e *Encoder) Close() {
	if e.filteredFrame != nil {
		ffmpeg.AVFrameFree(&e.filteredFrame)
	}
	if e.encPkt != nil {
		ffmpeg.AVPacketFree(&e.encPkt)
	}
	if e.decFrame != nil {
		ffmpeg.AVFrameFree(&e.decFrame)
	}
	if e.filterGraph != nil {
		ffmpeg.AVFilterGraphFree(&e.filterGraph)
	}
	if e.encCtx != nil {
		ffmpeg.AVCodecFreeContext(&e.encCtx)
	}
	if e.decCtx != nil {
		ffmpeg.AVCodecFreeContext(&e.decCtx)
	}
	if e.ofmtCtx != nil {
		if e.ofmtCtx.Oformat().Flags()&ffmpeg.AVFmtNofile == 0 && e.ofmtCtx.Pb() != nil {
			ffmpeg.AVIOClose(e.ofmtCtx.Pb())
			e.ofmtCtx.SetPb(nil)
		}
		ffmpeg.AVFormatFreeContext(e.ofmtCtx)
	}
	if e.ifmtCtx != nil {
		ffmpeg.AVFormatCloseInput(&e.ifmtCtx)
	}
}

// GetInputInfo returns information about the input audio
func (e *Encoder) GetInputInfo() (sampleRate, channels int, format string) {
	if e.decCtx == nil {
		return 0, 0, "unknown"
	}

	codecName := e.decCtx.Codec().Name()
	return e.decCtx.SampleRate(), e.decCtx.Channels(), codecName.String()
}
