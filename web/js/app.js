// We'll use React without JSX to avoid setting up Webpack and Babel.
// This is not supposed to be used as production code.

// Helper function to create React elements
const e = (el, props, children) => {
  if (props) {
    const { cls, ...rest } = props;
    return React.createElement(el, { ...rest, className: cls }, children);
  } else {
    return React.createElement(el, null, children);
  }
}

// Authentication API functions
function checkAuthStatus() {
  return fetch('/auth/status')
    .then(res => res.json())
    .catch(() => ({ authenticated: false }));
}

function login(username, password) {
  const formData = new URLSearchParams();
  formData.append('username', username);
  formData.append('password', password);

  return fetch('/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: formData
  }).then(res => res.json());
}

function logout() {
  return fetch('/logout', { method: 'POST' })
    .then(res => res.json());
}

function startSession(offer, language = 'auto', enableTranscribe = true) {
  return fetch('/session', {
    method: 'POST',
    body: JSON.stringify({
      offer,
      language,  // Pass language to server
      transcribe: enableTranscribe  // Whether to transcribe or just record
    }),
    headers: {
      'Content-Type': 'application/json'
    }
  }).then(res => {
    if (res.status === 401) {
      throw new Error('Unauthorized');
    }
    return res.json();
  }).then(msg => {
    return msg.answer;
  });
}

// Transcribe existing audio files
function transcribeFiles(files, language = 'auto') {
  return fetch('/transcribe', {
    method: 'POST',
    body: JSON.stringify({
      files,
      language
    }),
    headers: {
      'Content-Type': 'application/json'
    }
  }).then(res => {
    if (res.status === 401) {
      throw new Error('Unauthorized');
    }
    return res.json();
  });
}

// Handle different evt.data types according to the browser
function decodeDataChannelPayload(data) {
  if (data instanceof ArrayBuffer) {
    const dec = new TextDecoder('utf-8');
    return Promise.resolve(dec.decode(data));
  } else if (data instanceof Blob) {
    const reader = new FileReader();
    const readPromise = new Promise((accept, reject) => {
      reader.onload = () => accept(reader.result);
      reader.onerror = reject;
    });
    reader.readAsText(data, 'utf-8');
    return readPromise;
  }
}

function setupPeerConnection({ stream, onResult, onSignaling, onStop, language = 'auto', enableTranscribe = true }) {
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  });
  const resChan = pc.createDataChannel('results', {
    ordered: true,
    protocol: 'tcp'
  });
  resChan.onmessage = evt => {
    // evt.data will be an instance of ArrayBuffer OR Blob
    decodeDataChannelPayload(evt.data).then(strData => {
      const result = JSON.parse(strData);
      onResult(result);
    });
  };

  // We close everything when the data channel closes
  resChan.onclose = () => {
    pc.close();
    onStop()
  };

  pc.onicecandidate = evt => {
    if (!evt.candidate) {
      // ICE Gathering finished 
      const { sdp: offer } = pc.localDescription;
      startSession(offer, language, enableTranscribe).then(answer => {
        onSignaling(offer, answer);
        const rd = new RTCSessionDescription({
          sdp: answer,
          type: 'answer'
        });
        pc.setRemoteDescription(rd);
      });
    }
  };

  const audioTracks = stream.getAudioTracks();
  if (audioTracks.length > 0) {
    pc.addTrack(audioTracks[0], stream);
  }
  // Let's trigger ICE gathering
  pc.createOffer({
    offerToReceiveAudio: false,
    offerToReceiveVideo: false
  }).then(ld => {
    pc.setLocalDescription(ld)
  });
  return pc;
}

function LoadingSpinner() {
  return e('div', { 
    style: {
      display: 'inline-block',
      width: '20px',
      height: '20px',
      marginLeft: '10px',
      verticalAlign: 'middle',
      border: '3px solid rgba(0, 0, 0, 0.1)',
      borderRadius: '50%',
      borderTopColor: '#3273dc',
      animation: 'spin 1s ease-in-out infinite'
    }
  });
}

function ActionButton({ disabled, action, active, processing }) {
  // Add keyframes for spin animation if not already present
  if (!document.getElementById('spin-style')) {
    const style = document.createElement('style');
    style.id = 'spin-style';
    style.innerHTML = `
      @keyframes spin {
        to { transform: rotate(360deg); }
      }
      @keyframes pulse {
        0%, 100% { transform: scale(1); }
        50% { transform: scale(1.05); }
      }
    `;
    document.head.appendChild(style);
  }

  let buttonStyle;
  let buttonContent;
  let buttonDisabled = disabled;

  if (processing) {
    // Processing state - after stop, waiting for result
    buttonStyle = {
      background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
      border: 'none',
      color: 'white',
      borderRadius: '8px',
      padding: '0 1.5rem',
      fontWeight: '600',
      height: '40px',
      cursor: 'not-allowed'
    };
    buttonContent = [
      e('span', { cls: 'icon' }, 
        e('i', { cls: 'fas fa-cog fa-spin' })
      ),
      e('span', null, 'Processing...')
    ];
    buttonDisabled = true;
  } else if (active) {
    // Recording state
    buttonStyle = {
      background: 'linear-gradient(135deg, #ff416c 0%, #ff4b2b 100%)',
      border: 'none',
      color: 'white',
      borderRadius: '8px',
      padding: '0 1.5rem',
      fontWeight: '600',
      height: '40px',
      animation: 'pulse 2s infinite'
    };
    buttonContent = [
      e('span', { cls: 'icon' }, 
        e('i', { cls: 'fas fa-stop' })
      ),
      e('span', null, 'Stop'),
      e(LoadingSpinner)
    ];
  } else {
    // Idle state
    buttonStyle = {
      background: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)',
      border: 'none',
      color: 'white',
      borderRadius: '8px',
      padding: '0 1.5rem',
      fontWeight: '600',
      height: '40px'
    };
    buttonContent = [
      e('span', { cls: 'icon' }, 
        e('i', { cls: 'fas fa-microphone' })
      ),
      e('span', null, 'Start')
    ];
  }

  return e('button', {
    cls: 'button',
    onClick: action,
    disabled: buttonDisabled,
    style: buttonStyle
  }, buttonContent);
}

// Processing row component - shows in the table while transcribing
function ProcessingRow({ index, selectable }) {
  return e('tr', { style: { background: 'linear-gradient(90deg, #f0fdfa 0%, #e0f2fe 50%, #f0fdfa 100%)', animation: 'pulse 2s ease-in-out infinite' } }, [
    selectable && e('td', { style: { verticalAlign: 'middle' } }, '-'),
    e('th', { style: { verticalAlign: 'middle' } }, index + 1),
    e('td', { style: { verticalAlign: 'middle' } }, [
      e('div', { style: { display: 'flex', alignItems: 'center', gap: '12px' } }, [
        e('div', {
          style: {
            width: '24px',
            height: '24px',
            border: '3px solid #e0e0e0',
            borderTopColor: '#0891b2',
            borderRadius: '50%',
            animation: 'spin 1s linear infinite'
          }
        }),
        e('div', null, [
          e('div', { style: { fontWeight: '500', color: '#0891b2' } }, 'Transcribing audio...'),
          e('div', { cls: 'is-size-7 has-text-grey' }, 'This may take a moment with larger models')
        ])
      ])
    ]),
    e('td', { style: { verticalAlign: 'middle' } }, 
      e('span', { cls: 'has-text-grey-light' }, 'Processing...')
    ),
    e('td', { style: { verticalAlign: 'middle' } }, 
      e('span', { cls: 'has-text-grey-light' }, 'Processing...')
    ),
    e('td', { style: { verticalAlign: 'middle' } }, 
      e('span', { cls: 'has-text-grey-light' }, '-')
    )
  ]);
}

// Delete file API function
function deleteFile(filename) {
  return fetch(`/delete/${filename}`, {
    method: 'DELETE'
  }).then(res => res.json());
}

// Fetch text file content (first 100 words)
function fetchTextPreview(url) {
  return fetch(url)
    .then(res => res.text())
    .then(text => {
      const words = text.split(/\s+/).slice(0, 100);
      const preview = words.join(' ');
      return preview + (text.split(/\s+/).length > 100 ? '...' : '');
    })
    .catch(() => null);
}

// Custom Audio Player with visible progress bar
function AudioPlayer({ src, fileName }) {
  const audioRef = React.useRef(null);
  const [isPlaying, setIsPlaying] = React.useState(false);
  const [currentTime, setCurrentTime] = React.useState(0);
  const [duration, setDuration] = React.useState(0);
  const [isDragging, setIsDragging] = React.useState(false);

  const formatTime = (time) => {
    if (isNaN(time) || time === 0) return '0:00';
    const mins = Math.floor(time / 60);
    const secs = Math.floor(time % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const handlePlayPause = () => {
    if (audioRef.current) {
      if (isPlaying) {
        audioRef.current.pause();
      } else {
        audioRef.current.play();
      }
      setIsPlaying(!isPlaying);
    }
  };

  const handleTimeUpdate = () => {
    if (audioRef.current && !isDragging) {
      setCurrentTime(audioRef.current.currentTime);
    }
  };

  const handleLoadedMetadata = () => {
    if (audioRef.current) {
      setDuration(audioRef.current.duration);
    }
  };

  const handleEnded = () => {
    setIsPlaying(false);
    setCurrentTime(0);
  };

  const handleProgressClick = (evt) => {
    const progressBar = evt.currentTarget;
    const rect = progressBar.getBoundingClientRect();
    const clickX = evt.clientX - rect.left;
    const percentage = clickX / rect.width;
    const newTime = percentage * duration;
    
    if (audioRef.current) {
      audioRef.current.currentTime = newTime;
      setCurrentTime(newTime);
    }
  };

  const handleProgressDrag = (evt) => {
    if (!isDragging) return;
    handleProgressClick(evt);
  };

  const progressPercent = duration > 0 ? (currentTime / duration) * 100 : 0;

  return e('div', { style: { width: '100%', maxWidth: '280px' } }, [
    // Hidden audio element
    e('audio', {
      ref: audioRef,
      src: src,
      onTimeUpdate: handleTimeUpdate,
      onLoadedMetadata: handleLoadedMetadata,
      onEnded: handleEnded,
      preload: 'metadata'
    }),
    
    // File name tag
    fileName && e('div', { cls: 'file-tag mb-2', style: { display: 'inline-block' } }, [
      e('span', { cls: 'icon is-small' }, e('i', { cls: 'fas fa-file-audio' })),
      e('span', null, ` ${fileName}`)
    ]),
    
    // Player controls
    e('div', { 
      style: { 
        background: 'linear-gradient(135deg, #f0fdfa 0%, #e0f2fe 100%)',
        borderRadius: '12px',
        padding: '10px 12px',
        border: '1px solid #99f6e4'
      } 
    }, [
      // Progress bar container
      e('div', {
        style: {
          width: '100%',
          height: '8px',
          background: '#e0e0e0',
          borderRadius: '4px',
          cursor: 'pointer',
          position: 'relative',
          marginBottom: '8px'
        },
        onClick: handleProgressClick,
        onMouseDown: () => setIsDragging(true),
        onMouseUp: () => setIsDragging(false),
        onMouseLeave: () => setIsDragging(false),
        onMouseMove: handleProgressDrag
      }, [
        // Progress fill
        e('div', {
          style: {
            width: `${progressPercent}%`,
            height: '100%',
            background: 'linear-gradient(90deg, #0891b2 0%, #0d9488 100%)',
            borderRadius: '4px',
            transition: isDragging ? 'none' : 'width 0.1s ease'
          }
        }),
        // Draggable handle
        e('div', {
          style: {
            position: 'absolute',
            left: `calc(${progressPercent}% - 6px)`,
            top: '-4px',
            width: '16px',
            height: '16px',
            background: '#0891b2',
            borderRadius: '50%',
            boxShadow: '0 2px 6px rgba(8, 145, 178, 0.4)',
            cursor: 'grab',
            transition: isDragging ? 'none' : 'left 0.1s ease'
          }
        })
      ]),
      
      // Controls row
      e('div', { 
        style: { 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'space-between',
          gap: '8px'
        } 
      }, [
        // Play/Pause button
        e('button', {
          cls: 'button is-small',
          onClick: handlePlayPause,
          style: {
            background: isPlaying ? 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)' : 'linear-gradient(135deg, #0891b2 0%, #0d9488 100%)',
            color: 'white',
            border: 'none',
            borderRadius: '50%',
            width: '32px',
            height: '32px',
            padding: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }
        }, e('i', { cls: `fas ${isPlaying ? 'fa-pause' : 'fa-play'}`, style: { fontSize: '12px', marginLeft: isPlaying ? 0 : '2px' } })),
        
        // Time display
        e('span', { 
          style: { 
            fontSize: '0.75rem', 
            fontFamily: 'monospace',
            color: '#0f766e',
            minWidth: '80px',
            textAlign: 'center'
          } 
        }, `${formatTime(currentTime)} / ${formatTime(duration)}`)
      ])
    ])
  ]);
}

// Single result row component with text preview and delete
function ResultRow({ result, index, onDelete, selectable, selected, onSelect }) {
  const [textPreview, setTextPreview] = React.useState(null);
  const [loadingPreview, setLoadingPreview] = React.useState(false);
  const [deleting, setDeleting] = React.useState(false);
  const [showConfirm, setShowConfirm] = React.useState(false);

  // Helper to extract filename from path
  const getFileName = (path) => {
    if (!path) return null;
    return path.split(/[/\\]/).pop();
  };

  // Generate valid URLs for files if they exist
  const getFileUrl = (path) => {
    if (!path) return null;
    const filename = getFileName(path);
    return `/recordings/${filename}`;
  };

  const audioUrl = getFileUrl(result.audio_file);
  const textUrl = getFileUrl(result.text_file);
  const audioFileName = getFileName(result.audio_file);
  const textFileName = getFileName(result.text_file);
  
  // Check if this file has audio but no transcription (can be transcribed)
  const canTranscribe = audioFileName && !textFileName;

  // Fetch text preview when component mounts
  React.useEffect(() => {
    if (textUrl && !textPreview && !loadingPreview) {
      setLoadingPreview(true);
      fetchTextPreview(textUrl)
        .then(preview => {
          setTextPreview(preview);
          setLoadingPreview(false);
        })
        .catch(() => setLoadingPreview(false));
    }
  }, [textUrl]);

  const handleDelete = () => {
    setDeleting(true);
    const deletePromises = [];
    
    if (audioFileName) {
      deletePromises.push(deleteFile(audioFileName));
    }
    if (textFileName) {
      deletePromises.push(deleteFile(textFileName));
    }

    Promise.all(deletePromises)
      .then(() => {
        setDeleting(false);
        setShowConfirm(false);
        onDelete(index);
      })
      .catch(err => {
        console.error('Delete failed:', err);
        setDeleting(false);
        setShowConfirm(false);
      });
  };

  return e('tr', { style: selected ? { background: '#f0fdfa' } : {} }, [
    // Checkbox column (only show if selectable)
    selectable && e('td', { style: { verticalAlign: 'middle', width: '40px', textAlign: 'center' } }, 
      canTranscribe 
        ? e('input', {
            type: 'checkbox',
            checked: selected ? true : false,  // Explicit true/false
            onChange: (ev) => {
              onSelect(audioFileName);
            },
            style: { 
              width: '18px', 
              height: '18px', 
              cursor: 'pointer',
              accentColor: '#0891b2'
            }
          })
        : e('span', { 
            title: 'Already transcribed - no action needed',
            style: { 
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '18px', 
              height: '18px',
              background: '#d1fae5',
              color: '#059669',
              borderRadius: '4px',
              fontSize: '12px',
              fontWeight: 'bold'
            }
          }, '✓')
    ),
    e('th', { style: { verticalAlign: 'middle' } }, index + 1),
    e('td', { style: { verticalAlign: 'middle', maxWidth: '400px' } }, [
      // Show text preview if available
      textPreview 
        ? e('div', { 
            style: { 
              fontSize: '0.9rem', 
              lineHeight: '1.5',
              maxHeight: '100px',
              overflow: 'hidden',
              textOverflow: 'ellipsis'
            } 
          }, textPreview)
        : loadingPreview
          ? e('span', { cls: 'has-text-grey is-size-7' }, 'Loading preview...')
          : canTranscribe
            ? e('div', { style: { color: '#9ca3af', fontStyle: 'italic' } }, 'Not transcribed yet')
            : e('div', { style: { fontWeight: '500' } }, result.text),
      !canTranscribe && e('div', { cls: 'is-size-7 has-text-grey', style: { marginTop: '4px' } }, 
        `Confidence: ${(result.confidence * 100).toFixed(1)}%`
      )
    ]),
    e('td', { style: { verticalAlign: 'middle' } }, audioUrl 
      ? e('div', null, [
          e(AudioPlayer, { src: audioUrl, fileName: audioFileName }),
          e('a', { href: audioUrl, download: '', cls: 'button is-small is-info is-light mt-2' }, [
            e('span', { cls: 'icon is-small' }, e('i', { cls: 'fas fa-download' })),
            e('span', null, 'Download')
          ])
        ])
      : e('span', { cls: 'has-text-grey-light' }, '-')
    ),
    e('td', { style: { verticalAlign: 'middle' } }, textUrl 
      ? e('div', null, [
          textFileName && e('div', { cls: 'file-tag mb-2', style: { display: 'inline-block' } }, [
            e('span', { cls: 'icon is-small' }, e('i', { cls: 'fas fa-file-lines' })),
            e('span', null, ` ${textFileName}`)
          ]),
          e('div', { cls: 'buttons are-small' }, [
            e('a', { href: textUrl, target: '_blank', cls: 'button is-link is-light' }, [
              e('span', { cls: 'icon is-small' }, e('i', { cls: 'fas fa-eye' })),
              e('span', null, 'View')
            ]),
            e('a', { href: textUrl, download: '', cls: 'button is-primary is-light' }, [
              e('span', { cls: 'icon is-small' }, e('i', { cls: 'fas fa-download' })),
              e('span', null, 'Download')
            ])
          ])
        ])
      : e('span', { cls: 'has-text-grey-light' }, '-')
    ),
    e('td', { style: { verticalAlign: 'middle' } }, [
      showConfirm 
        ? e('div', null, [
            e('p', { cls: 'is-size-7 mb-2' }, 'Delete files?'),
            e('div', { cls: 'buttons are-small' }, [
              e('button', { 
                cls: `button is-danger is-small ${deleting ? 'is-loading' : ''}`,
                onClick: handleDelete,
                disabled: deleting
              }, 'Yes'),
              e('button', { 
                cls: 'button is-light is-small',
                onClick: () => setShowConfirm(false),
                disabled: deleting
              }, 'No')
            ])
          ])
        : e('button', { 
            cls: 'button is-danger is-light is-small',
            onClick: () => setShowConfirm(true)
          }, [
            e('span', { cls: 'icon is-small' }, e('i', { cls: 'fas fa-trash' })),
            e('span', null, 'Delete')
          ])
    ])
  ]);
}

function Results({ results, onDeleteResult, processing, selectable, selectedFiles, onSelectFile }) {
  // Show empty state only if no results AND not processing
  if (!results.length && !processing) {
    return e('div', { cls: 'has-text-centered', style: { padding: '3rem' } }, [
      e('span', { cls: 'icon is-large has-text-grey-light' }, 
        e('i', { cls: 'fas fa-microphone-slash fa-3x' })
      ),
      e('p', { cls: 'has-text-grey mt-3' }, 'No results yet. Click Start to begin recording.')
    ]);
  }

  // Check if there are any untranscribed files (for showing info message)
  const untranscribedCount = results.filter(r => {
    const audioFileName = r.audio_file ? r.audio_file.split(/[/\\]/).pop() : null;
    const textFileName = r.text_file ? r.text_file.split(/[/\\]/).pop() : null;
    return audioFileName && !textFileName;
  }).length;

  return e('div', null, [
    // Show info message when in selectable mode but all files are already transcribed
    selectable && untranscribedCount === 0 && e('div', { 
      cls: 'notification is-info is-light mb-4',
      style: { 
        display: 'flex', 
        alignItems: 'center', 
        gap: '10px',
        background: 'linear-gradient(135deg, #e0f2fe 0%, #f0fdfa 100%)',
        border: '1px solid #0891b2',
        borderRadius: '8px'
      }
    }, [
      e('span', { cls: 'icon' }, e('i', { cls: 'fas fa-info-circle', style: { color: '#0891b2' } })),
      e('span', null, [
        e('strong', null, 'All files are already transcribed. '),
        'The ✓ symbol indicates the file has been transcribed. Record new audio to transcribe.'
      ])
    ]),
    
    // Show count of selectable files when in selectable mode
    selectable && untranscribedCount > 0 && e('div', { 
      cls: 'notification is-success is-light mb-4',
      style: { 
        display: 'flex', 
        alignItems: 'center', 
        gap: '10px',
        background: 'linear-gradient(135deg, #d1fae5 0%, #f0fdfa 100%)',
        border: '1px solid #10b981',
        borderRadius: '8px'
      }
    }, [
      e('span', { cls: 'icon' }, e('i', { cls: 'fas fa-check-circle', style: { color: '#10b981' } })),
      e('span', null, `${untranscribedCount} file(s) available for transcription. Select the files you want to transcribe.`)
    ]),
    
    e('div', { cls: 'table-container' }, 
      e('table', { cls: 'table is-fullwidth is-hoverable' }, [
        e('thead', null, 
          e('tr', { style: { background: 'linear-gradient(135deg, #0891b2 0%, #0d9488 100%)' } }, [
            selectable && e('th', { style: { width: '40px', color: 'white', textAlign: 'center' } }, 
              e('span', { title: 'Select files to transcribe' }, e('i', { cls: 'fas fa-check-square' }))
            ),
            e('th', { style: { width: '50px', color: 'white' } }, '#'),
            e('th', { style: { color: 'white' } }, 'Transcription'),
            e('th', { style: { width: '280px', color: 'white' } }, 'Audio File'),
            e('th', { style: { width: '200px', color: 'white' } }, 'Text File'),
            e('th', { style: { width: '100px', color: 'white' } }, 'Actions')
          ])
        ),
        e('tbody', null, [
          // Show processing row at the top when transcribing
          processing && e(ProcessingRow, { key: 'processing', index: results.length, selectable }),
          // Then show existing results
          ...results.map((r, i) => {
            const fileName = r.audio_file ? r.audio_file.split(/[/\\]/).pop() : null;
            const isSelected = fileName && selectedFiles && selectedFiles.includes(fileName);
            return e(ResultRow, { 
              key: i, 
              result: r, 
              index: i, 
              onDelete: onDeleteResult,
              selectable,
              selected: isSelected,
              onSelect: onSelectFile
            });
          })
        ])
      ])
    )
  ]);
}

function Waveform({ stream }) {
  const canvasRef = React.useRef(null);

  React.useEffect(() => {
    if (!stream || !canvasRef.current) return;

    const AudioContext = window.AudioContext || window.webkitAudioContext;
    const audioCtx = new AudioContext();
    const analyser = audioCtx.createAnalyser();
    const source = audioCtx.createMediaStreamSource(stream);
    source.connect(analyser);

    analyser.fftSize = 2048;
    const bufferLength = analyser.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);
    const canvas = canvasRef.current;
    const canvasCtx = canvas.getContext('2d');

    let animationId;

    const draw = () => {
      animationId = requestAnimationFrame(draw);

      analyser.getByteTimeDomainData(dataArray);

      // Create gradient background
      const gradient = canvasCtx.createLinearGradient(0, 0, 0, canvas.height);
      gradient.addColorStop(0, '#f8fafc');
      gradient.addColorStop(1, '#e2e8f0');
      canvasCtx.fillStyle = gradient;
      canvasCtx.fillRect(0, 0, canvas.width, canvas.height);

      // Create gradient for waveform
      const waveGradient = canvasCtx.createLinearGradient(0, 0, canvas.width, 0);
      waveGradient.addColorStop(0, '#0891b2');
      waveGradient.addColorStop(0.5, '#0d9488');
      waveGradient.addColorStop(1, '#0891b2');

      canvasCtx.lineWidth = 3;
      canvasCtx.strokeStyle = waveGradient;
      canvasCtx.lineCap = 'round';
      canvasCtx.lineJoin = 'round';

      canvasCtx.beginPath();

      const sliceWidth = canvas.width * 1.0 / bufferLength;
      let x = 0;

      for (let i = 0; i < bufferLength; i++) {
        const v = dataArray[i] / 128.0;
        const y = v * canvas.height / 2;

        if (i === 0) {
          canvasCtx.moveTo(x, y);
        } else {
          canvasCtx.lineTo(x, y);
        }

        x += sliceWidth;
      }

      canvasCtx.lineTo(canvas.width, canvas.height / 2);
      canvasCtx.stroke();
    };

    draw();

    return () => {
      cancelAnimationFrame(animationId);
      audioCtx.close();
    };
  }, [stream]);

  return e('div', { style: { marginBottom: '1.5rem', marginTop: '1rem' } }, [
    e('h4', { cls: 'title is-6', style: { display: 'flex', alignItems: 'center', marginBottom: '0.75rem' } }, [
      e('span', { cls: 'icon mr-2', style: { color: '#0891b2' } }, e('i', { cls: 'fas fa-wave-square' })),
      'Live Audio Waveform'
    ]),
    e('canvas', { 
      ref: canvasRef, 
      width: 800, 
      height: 120, 
      style: { 
        width: '100%', 
        height: '120px', 
        borderRadius: '12px',
        boxShadow: 'inset 0 2px 8px rgba(0,0,0,0.1)'
      } 
    })
  ]);
}

function DeviceSelector({ devices, selectedDeviceId, onSelect, disabled, showLabel = false }) {
  return e('div', { cls: 'field', style: { marginBottom: 0 } }, [
    showLabel && e('label', { cls: 'label' }, 'Audio Input Device'),
    e('div', { cls: 'control' }, [
      e('div', { cls: 'select is-fullwidth' }, [
        e('select', { 
          onChange: (evt) => onSelect(evt.target.value),
          value: selectedDeviceId,
          disabled,
          style: { height: '40px' }
        }, [
          e('option', { value: '' }, 'Select Audio Device...'),
          ...devices.map(d => 
            e('option', { key: d.deviceId, value: d.deviceId }, d.label || `Microphone ${d.deviceId.slice(0, 5)}...`)
          )
        ])
      ])
    ])
  ]);
}

// Supported languages for Whisper
const SUPPORTED_LANGUAGES = [
  { code: 'auto', name: 'Auto Detect', flag: '🌐' },
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'zh', name: 'Chinese (中文)', flag: '🇨🇳' },
  { code: 'ja', name: 'Japanese (日本語)', flag: '🇯🇵' },
  { code: 'ko', name: 'Korean (한국어)', flag: '🇰🇷' },
  { code: 'es', name: 'Spanish (Español)', flag: '🇪🇸' },
  { code: 'fr', name: 'French (Français)', flag: '🇫🇷' },
  { code: 'de', name: 'German (Deutsch)', flag: '🇩🇪' },
  { code: 'it', name: 'Italian (Italiano)', flag: '🇮🇹' },
  { code: 'pt', name: 'Portuguese (Português)', flag: '🇵🇹' },
  { code: 'ru', name: 'Russian (Русский)', flag: '🇷🇺' },
  { code: 'ar', name: 'Arabic (العربية)', flag: '🇸🇦' },
  { code: 'hi', name: 'Hindi (हिन्दी)', flag: '🇮🇳' },
  { code: 'th', name: 'Thai (ไทย)', flag: '🇹🇭' },
  { code: 'vi', name: 'Vietnamese (Tiếng Việt)', flag: '🇻🇳' },
  { code: 'nl', name: 'Dutch (Nederlands)', flag: '🇳🇱' },
  { code: 'pl', name: 'Polish (Polski)', flag: '🇵🇱' },
  { code: 'tr', name: 'Turkish (Türkçe)', flag: '🇹🇷' },
  { code: 'sv', name: 'Swedish (Svenska)', flag: '🇸🇪' },
  { code: 'id', name: 'Indonesian (Bahasa)', flag: '🇮🇩' }
];

function LanguageSelector({ selectedLanguage, onSelect, disabled }) {
  return e('div', { cls: 'field', style: { marginBottom: 0 } }, [
    e('div', { cls: 'control' }, [
      e('div', { cls: 'select', style: { width: '180px' } }, [
        e('select', { 
          onChange: (evt) => onSelect(evt.target.value),
          value: selectedLanguage,
          disabled,
          style: { height: '40px' }
        }, SUPPORTED_LANGUAGES.map(lang => 
          e('option', { key: lang.code, value: lang.code }, `${lang.flag} ${lang.name}`)
        ))
      ])
    ])
  ]);
}

function SessionStats({ duration, stats }) {
  const mins = Math.floor(duration / 60).toString().padStart(2, '0');
  const secs = (duration % 60).toString().padStart(2, '0');

  return e('nav', { cls: 'level is-mobile box' }, [
    e('div', { cls: 'level-item has-text-centered' }, [
      e('div', null, [
        e('p', { cls: 'heading' }, 'Time'),
        e('p', { cls: 'title is-5' }, `${mins}:${secs}`)
      ])
    ]),
    e('div', { cls: 'level-item has-text-centered' }, [
      e('div', null, [
        e('p', { cls: 'heading' }, 'Codec'),
        e('p', { cls: 'title is-5' }, stats.codec || '-')
      ])
    ]),
    e('div', { cls: 'level-item has-text-centered' }, [
      e('div', null, [
        e('p', { cls: 'heading' }, 'Transport'),
        e('p', { cls: 'title is-5' }, stats.transport || '-')
      ])
    ])
  ]);
}

function CollapsibleSection({ title, content }) {
  const [isOpen, setIsOpen] = React.useState(false);

  return e('div', { style: { marginBottom: '1rem' } }, [
    e('div', { 
      onClick: () => setIsOpen(!isOpen),
      style: { cursor: 'pointer', display: 'flex', alignItems: 'center' }
    }, [
      e('span', { style: { marginRight: '0.5rem', transform: isOpen ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.2s' } }, '▶'),
      e('h3', { cls: 'subtitle', style: { marginBottom: 0 } }, title)
    ]),
    isOpen && e('pre', { cls: 'is-family-code', style: { marginTop: '0.5rem' } }, content || '-')
  ]);
}

// Login Form Component
function LoginForm({ onLogin }) {
  const [username, setUsername] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(false);

  const handleSubmit = (evt) => {
    evt.preventDefault();
    setError('');
    setLoading(true);

    login(username, password)
      .then(result => {
        setLoading(false);
        if (result.success) {
          onLogin(result.username);
        } else {
          setError(result.message || 'Login failed');
        }
      })
      .catch(err => {
        setLoading(false);
        setError('Network error. Please try again.');
      });
  };

  return e('div', { 
    cls: 'card-elevated', 
    style: { 
      maxWidth: '420px', 
      margin: '80px auto', 
      padding: '2.5rem',
      background: 'white'
    } 
  }, [
    e('div', { cls: 'has-text-centered mb-5' }, [
      e('span', { 
        cls: 'icon is-large', 
        style: { 
          background: 'linear-gradient(135deg, #0891b2 0%, #0d9488 100%)',
          borderRadius: '50%',
          width: '80px',
          height: '80px',
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center'
        }
      }, e('i', { cls: 'fas fa-microphone fa-2x', style: { color: 'white' } })),
      e('h2', { cls: 'title is-4 mt-4 title-gradient' }, 'Welcome Back'),
      e('p', { cls: 'has-text-grey' }, 'Sign in to continue')
    ]),
    e('form', { onSubmit: handleSubmit }, [
      e('div', { cls: 'field' }, [
        e('label', { cls: 'label', style: { fontWeight: '500' } }, 'Username'),
        e('div', { cls: 'control has-icons-left' }, [
          e('input', {
            cls: 'input is-medium',
            type: 'text',
            placeholder: 'Enter username',
            value: username,
            onChange: (evt) => setUsername(evt.target.value),
            required: true,
            style: { borderRadius: '8px' }
          }),
          e('span', { cls: 'icon is-small is-left' }, 
            e('i', { cls: 'fas fa-user', style: { color: '#0891b2' } })
          )
        ])
      ]),
      e('div', { cls: 'field' }, [
        e('label', { cls: 'label', style: { fontWeight: '500' } }, 'Password'),
        e('div', { cls: 'control has-icons-left' }, [
          e('input', {
            cls: 'input is-medium',
            type: 'password',
            placeholder: 'Enter password',
            value: password,
            onChange: (evt) => setPassword(evt.target.value),
            required: true,
            style: { borderRadius: '8px' }
          }),
          e('span', { cls: 'icon is-small is-left' }, 
            e('i', { cls: 'fas fa-lock', style: { color: '#0891b2' } })
          )
        ])
      ]),
      error && e('div', { cls: 'notification is-danger is-light', style: { borderRadius: '8px' } }, error),
      e('div', { cls: 'field mt-5' }, [
        e('button', { 
          cls: `button is-gradient is-fullwidth is-medium ${loading ? 'is-loading' : ''}`,
          type: 'submit',
          disabled: loading,
          style: { borderRadius: '8px', fontWeight: '600' }
        }, [
          e('span', { cls: 'icon' }, e('i', { cls: 'fas fa-sign-in-alt' })),
          e('span', null, 'Sign In')
        ])
      ])
    ])
  ]);
}

// Navbar Component (shows welcome message and logout button on top right)
function Navbar({ username, onLogout }) {
  const handleLogout = () => {
    logout().then(() => onLogout());
  };

  return e('nav', { 
    cls: 'navbar is-fixed-top navbar-custom',
    role: 'navigation',
    style: { padding: '0 1rem' }
  }, [
    e('div', { cls: 'container' }, [
      e('div', { cls: 'navbar-brand' }, [
        e('a', { cls: 'navbar-item', href: '/', style: { display: 'flex', alignItems: 'center' } }, [
          e('span', { 
            style: { 
              background: 'linear-gradient(135deg, #0891b2 0%, #0d9488 100%)',
              borderRadius: '12px',
              width: '42px',
              height: '42px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              marginRight: '12px',
              boxShadow: '0 4px 12px rgba(8, 145, 178, 0.3)'
            }
          }, e('i', { cls: 'fas fa-couch', style: { color: 'white', fontSize: '1.2rem' } })),
          e('span', { 
            cls: 'has-text-weight-bold', 
            style: { 
              background: 'linear-gradient(135deg, #0891b2 0%, #0d9488 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              backgroundClip: 'text',
              fontSize: '1.1rem'
            } 
          }, 'Lazy Speech To Text')
        ])
      ]),
      e('div', { cls: 'navbar-end' }, [
        e('div', { cls: 'navbar-item' }, [
          e('div', { cls: 'welcome-badge' }, [
            e('span', { cls: 'icon is-small mr-1' }, e('i', { cls: 'fas fa-user' })),
            e('span', null, `Welcome, ${username}`)
          ])
        ]),
        e('div', { cls: 'navbar-item' }, [
          e('button', { 
            cls: 'button is-light is-small',
            onClick: handleLogout,
            style: { borderRadius: '20px' }
          }, [
            e('span', { cls: 'icon is-small' }, e('i', { cls: 'fas fa-sign-out-alt' })),
            e('span', null, 'Logout')
          ])
        ])
      ])
    ])
  ]);
}

function AppContent() {
  const [state, setState] = React.useState(AppContent.initialState);

  // Fetch audio devices and existing files on mount
  React.useEffect(() => {
    // Enumerate audio devices
    navigator.mediaDevices.enumerateDevices()
      .then(devices => {
        const audioInputs = devices.filter(d => d.kind === 'audioinput');
        setState(st => ({ ...st, devices: audioInputs }));
      })
      .catch(err => console.error('Error enumerating devices:', err));

    // Fetch existing files
    fetch('/files')
      .then(res => res.json())
      .then(files => {
        // Group files by base name (without extension)
        // Files now have format: { name: "filename.wav", modTime: 1234567890 }
        const groups = {};
        files.forEach(fileInfo => {
          const f = fileInfo.name;
          const modTime = fileInfo.modTime;
          // Extract base name: "whisper_audio_1_20230101_120000.wav" -> "whisper_audio_1_20230101_120000"
          const baseName = f.substring(0, f.lastIndexOf('.'));
          const ext = f.substring(f.lastIndexOf('.') + 1);
          
          if (!groups[baseName]) {
            groups[baseName] = { text: baseName, confidence: 1.0, modTime: modTime }; // Default values
          }
          // Update modTime to the latest
          if (modTime > groups[baseName].modTime) {
            groups[baseName].modTime = modTime;
          }
          
          if (ext === 'wav') {
            groups[baseName].audio_file = `recordings/${f}`;
          } else if (ext === 'txt') {
            groups[baseName].text_file = `recordings/${f}`;
            // Try to load text content for preview? Maybe too heavy. 
            // We'll mark it as having a text file available.
            groups[baseName].text = "Transcription available";
          }
        });

        // Convert groups to results array
        const results = Object.values(groups)
          // Filter to show only if at least one file exists
          .filter(g => g.audio_file || g.text_file)
          // Sort by modification time descending (newest first)
          .sort((a, b) => b.modTime - a.modTime);

        setState(st => ({ ...st, results: [...results, ...st.results] }));
      })
      .catch(err => console.error('Error fetching files:', err));
  }, []);

  React.useEffect(() => {
    let intervalId;
    if (state.active && state.pc) {
      const startTime = Date.now();
      intervalId = setInterval(() => {
        const duration = Math.floor((Date.now() - startTime) / 1000);

        state.pc.getStats().then(statsReport => {
          let codec = '-';
          let transport = '-';

          statsReport.forEach(report => {
            if (report.type === 'outbound-rtp' && report.mediaType === 'audio') {
              const codecReport = statsReport.get(report.codecId);
              if (codecReport && codecReport.mimeType) {
                // e.g. "audio/opus" -> "opus"
                codec = codecReport.mimeType.split('/')[1] || codecReport.mimeType;
              }
            }
            if (report.type === 'candidate-pair' && report.state === 'succeeded') {
              const localCandidate = statsReport.get(report.localCandidateId);
              if (localCandidate && localCandidate.protocol) {
                transport = localCandidate.protocol.toUpperCase();
              }
            }
          });

          setState(st => ({
            ...st,
            recordingDuration: duration,
            stats: { codec, transport }
          }));
        });
      }, 1000);
    }
    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [state.active, state.pc]);

  function start() {
    setState(st => ({ ...st, offer: null, answer: null, error: null, recordingDuration: 0, stats: { codec: '-', transport: '-' } }));

    const audioConstraints = state.selectedDeviceId 
      ? { deviceId: { exact: state.selectedDeviceId } } 
      : true;

    navigator.mediaDevices.getUserMedia({
      audio: audioConstraints,
      video: false
    }).then(stream => {
      const pc = setupPeerConnection({
        stream, 
        language: state.selectedLanguage,  // Pass selected language
        enableTranscribe: state.enableTranscribe,  // Pass transcribe option
        onSignaling: (offer, answer) => setState(st => ({ ...st, offer, answer })),
        onResult: (r) => setState(st => ({ 
          ...st, 
          results: [...st.results, r],
          processing: false  // Result received, stop processing
        })),
        onStop: () => setState(st => ({ ...st, pc: null })),
      });

      setState(st => ({ ...st, stream, pc, active: true, processing: false }));
    }).catch(error => {
      setState(st => ({ ...st, error, processing: false }));
    });
  }

  function stop() {
    state.stream && state.stream.getAudioTracks().forEach(tr => tr.stop());
    // Set processing to true and record current results count
    setState(st => ({ 
      ...st, 
      stream: null, 
      active: false, 
      processing: true,
      resultsCountBeforeStop: st.results.length
    }));
  }

  // Watch for new results to turn off processing state
  React.useEffect(() => {
    if (state.processing && state.results.length > state.resultsCountBeforeStop) {
      // New result arrived, turn off processing
      setState(st => ({ ...st, processing: false }));
    }
  }, [state.results.length, state.processing, state.resultsCountBeforeStop]);

  // Auto-timeout for processing state (30 seconds max)
  React.useEffect(() => {
    if (state.processing) {
      const timeout = setTimeout(() => {
        setState(st => ({ ...st, processing: false }));
      }, 30000); // 30 second timeout
      return () => clearTimeout(timeout);
    }
  }, [state.processing]);

  // Auto-select first untranscribed file when switching to "Transcribe only" mode
  // Use a ref to track if we've already auto-selected to prevent repeated selections
  const hasAutoSelected = React.useRef(false);
  
  React.useEffect(() => {
    const isTranscribeOnlyMode = !state.enableRecord && state.enableTranscribe;
    
    if (isTranscribeOnlyMode && state.results.length > 0 && !hasAutoSelected.current) {
      // Find the first file that has audio but no transcription (sorted by modTime desc, so newest first)
      const firstUntranscribed = state.results.find(r => {
        const audioFileName = r.audio_file ? r.audio_file.split(/[/\\]/).pop() : null;
        const textFileName = r.text_file ? r.text_file.split(/[/\\]/).pop() : null;
        return audioFileName && !textFileName;
      });
      
      if (firstUntranscribed) {
        const fileName = firstUntranscribed.audio_file.split(/[/\\]/).pop();
        setState(st => ({ ...st, selectedFiles: [fileName] }));
        hasAutoSelected.current = true;
      }
    } else if (!isTranscribeOnlyMode) {
      // Clear selection and reset auto-select flag when not in transcribe-only mode
      if (state.selectedFiles.length > 0) {
        setState(st => ({ ...st, selectedFiles: [] }));
      }
      hasAutoSelected.current = false;
    }
  }, [state.enableRecord, state.enableTranscribe, state.results.length]);

  // Check if at least one option is selected
  const canStart = state.enableRecord || state.enableTranscribe;
  
  // Handle file selection for transcription
  const handleSelectFile = (fileName) => {
    setState(st => {
      const isSelected = st.selectedFiles.includes(fileName);
      return {
        ...st,
        selectedFiles: isSelected 
          ? st.selectedFiles.filter(f => f !== fileName)
          : [...st.selectedFiles, fileName]
      };
    });
  };

  // Handle transcribe selected files
  const handleTranscribeSelected = () => {
    if (state.selectedFiles.length === 0) return;
    
    setState(st => ({ ...st, transcribingFiles: true }));
    
    transcribeFiles(state.selectedFiles, state.selectedLanguage)
      .then(results => {
        // Update results with transcription data
        setState(st => {
          const updatedResults = st.results.map(r => {
            const fileName = r.audio_file ? r.audio_file.split(/[/\\]/).pop() : null;
            const transcribed = results.find(tr => tr.audio_file && tr.audio_file.includes(fileName));
            if (transcribed) {
              return { ...r, ...transcribed };
            }
            return r;
          });
          return {
            ...st,
            results: updatedResults,
            selectedFiles: [],
            transcribingFiles: false
          };
        });
      })
      .catch(err => {
        console.error('Transcription failed:', err);
        setState(st => ({ ...st, transcribingFiles: false }));
      });
  };

  const action = state.active ? stop: start;
  
  // Wrapper for start that checks options
  const handleAction = () => {
    if (!canStart && !state.active) {
      alert('Please select at least one option: Record or Transcribe');
      return;
    }
    action();
  };

  return e('div', null, [
    // Options checkboxes
    e('div', { 
      cls: 'box mb-4', 
      style: { 
        background: 'linear-gradient(135deg, #f0fdfa 0%, #e0f2fe 100%)',
        padding: '1rem 1.5rem',
        borderRadius: '12px'
      } 
    }, [
      e('div', { cls: 'field is-grouped' }, [
        e('div', { cls: 'control' }, [
          e('label', { cls: 'checkbox', style: { display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' } }, [
            e('input', {
              type: 'checkbox',
              checked: state.enableRecord,
              onChange: () => setState(st => ({ ...st, enableRecord: !st.enableRecord })),
              disabled: state.active || state.processing,
              style: { width: '18px', height: '18px', accentColor: '#0891b2' }
            }),
            e('span', { style: { fontWeight: '500', color: '#0f766e' } }, [
              e('i', { cls: 'fas fa-microphone mr-2' }),
              'Record Audio'
            ])
          ])
        ]),
        e('div', { cls: 'control ml-5' }, [
          e('label', { cls: 'checkbox', style: { display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' } }, [
            e('input', {
              type: 'checkbox',
              checked: state.enableTranscribe,
              onChange: () => setState(st => ({ ...st, enableTranscribe: !st.enableTranscribe })),
              disabled: state.active || state.processing,
              style: { width: '18px', height: '18px', accentColor: '#0891b2' }
            }),
            e('span', { style: { fontWeight: '500', color: '#0f766e' } }, [
              e('i', { cls: 'fas fa-language mr-2' }),
              'Transcribe Audio'
            ])
          ])
        ]),
        // Warning if nothing selected
        !canStart && e('div', { cls: 'control ml-4' }, [
          e('span', { cls: 'tag is-warning is-light' }, [
            e('i', { cls: 'fas fa-exclamation-triangle mr-2' }),
            'Select at least one option'
          ])
        ])
      ])
    ]),

    e('div', { cls: 'field is-grouped is-align-items-flex-end' }, [
      e('div', { cls: 'control' }, [
        e(ActionButton, { 
          active: state.active, 
          action: handleAction, 
          disabled: ((!!state.pc) && !state.active) || (!canStart && !state.active),
          processing: state.processing
        })
      ]),
      e('div', { cls: 'control is-expanded' }, [
        e(DeviceSelector, { 
          devices: state.devices, 
          selectedDeviceId: state.selectedDeviceId, 
          onSelect: (id) => setState(st => ({ ...st, selectedDeviceId: id })),
          disabled: state.active || state.processing
        })
      ]),
      e('div', { cls: 'control' }, [
        e(LanguageSelector, { 
          selectedLanguage: state.selectedLanguage, 
          onSelect: (lang) => setState(st => ({ ...st, selectedLanguage: lang })),
          disabled: state.active || state.processing
        })
      ])
    ]),
    
    // Display Recording Info if active or results exist
    (state.active || state.recordingDuration > 0) && e(SessionStats, { duration: state.recordingDuration, stats: state.stats }),

    state.stream && e(Waveform, { stream: state.stream }),
    
    e('div', { cls: 'mt-5' }, [
      e('div', { style: { display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem' } }, [
        e('h3', { cls: 'title is-5 mb-0', style: { display: 'flex', alignItems: 'center' } }, [
          e('span', { cls: 'icon mr-2', style: { color: '#0891b2' } }, e('i', { cls: 'fas fa-list-ul' })),
          'Transcription Results',
          // Show small processing indicator next to title when processing
          (state.processing || state.transcribingFiles) && e('span', { 
            cls: 'ml-3', 
            style: { 
              fontSize: '0.8rem', 
              color: '#0891b2',
              display: 'flex',
              alignItems: 'center',
              gap: '6px'
            } 
          }, [
            e('div', {
              style: {
                width: '14px',
                height: '14px',
                border: '2px solid #e0e0e0',
                borderTopColor: '#0891b2',
                borderRadius: '50%',
                animation: 'spin 1s linear infinite'
              }
            }),
            'Processing...'
          ])
        ]),
        // Transcribe selected button (only show when transcribe-only mode and files selected)
        !state.enableRecord && state.enableTranscribe && state.selectedFiles.length > 0 && e('button', {
          cls: `button is-gradient ${state.transcribingFiles ? 'is-loading' : ''}`,
          onClick: handleTranscribeSelected,
          disabled: state.transcribingFiles,
          style: { borderRadius: '8px' }
        }, [
          e('span', { cls: 'icon' }, e('i', { cls: 'fas fa-language' })),
          e('span', null, `Transcribe Selected (${state.selectedFiles.length})`)
        ])
      ]),
      e(Results, { 
        results: state.results,
        processing: state.processing || state.transcribingFiles,
        selectable: !state.enableRecord && state.enableTranscribe,
        selectedFiles: state.selectedFiles,
        onSelectFile: handleSelectFile,
        onDeleteResult: (index) => {
          setState(st => ({
            ...st,
            results: st.results.filter((_, i) => i !== index)
          }));
        }
      })
    ]),
    
    e('div', { cls: 'mt-5' }, [
      e(CollapsibleSection, { title: 'SDP Offer', content: state.offer }),
      e(CollapsibleSection, { title: 'SDP Answer', content: state.answer })
    ])
  ])
}

AppContent.initialState = {
  pc: null,
  stream: null,
  offer: null,
  answer: null,
  error: null,
  results: [],
  active: false,
  processing: false,  // New state for processing after stop
  devices: [],
  selectedDeviceId: '',
  selectedLanguage: 'auto',  // Default to auto-detect
  recordingDuration: 0,
  stats: { codec: '-', transport: '-' },
  resultsCountBeforeStop: 0,  // Track results count when stopping
  enableRecord: true,      // Record audio checkbox
  enableTranscribe: true,  // Transcribe audio checkbox
  selectedFiles: [],       // Selected files for transcription
  transcribingFiles: false // Processing selected files
};

function App() {
  const [authState, setAuthState] = React.useState({
    checking: true,
    authenticated: false,
    username: ''
  });

  // Check authentication status on mount
  React.useEffect(() => {
    checkAuthStatus().then(result => {
      setAuthState({
        checking: false,
        authenticated: result.authenticated,
        username: result.username || ''
      });
    });
  }, []);

  const handleLogin = (username) => {
    setAuthState({
      checking: false,
      authenticated: true,
      username
    });
  };

  const handleLogout = () => {
    setAuthState({
      checking: false,
      authenticated: false,
      username: ''
    });
  };

  // Show loading while checking auth
  if (authState.checking) {
    return e('section', { cls: 'section', style: { minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' } }, [
      e('div', { cls: 'has-text-centered' }, [
        e('div', { 
          style: { 
            width: '60px', 
            height: '60px', 
            border: '4px solid #e0e0e0',
            borderTopColor: '#0891b2',
            borderRadius: '50%',
            animation: 'spin 1s linear infinite',
            margin: '0 auto'
          }
        }),
        e('p', { cls: 'mt-4 has-text-grey' }, 'Loading...')
      ])
    ]);
  }

  // Show login form if not authenticated
  if (!authState.authenticated) {
    return e('section', { cls: 'section', style: { minHeight: '100vh' } }, [
      e('div', { cls: 'container' }, [
        e('div', { cls: 'has-text-centered mb-5' }, [
          e('div', { style: { marginBottom: '1.5rem' } }, [
            e('span', { 
              style: { 
                background: 'linear-gradient(135deg, #0891b2 0%, #0d9488 100%)',
                borderRadius: '20px',
                width: '80px',
                height: '80px',
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: '0 8px 24px rgba(8, 145, 178, 0.3)'
              }
            }, e('i', { cls: 'fas fa-couch fa-2x', style: { color: 'white' } }))
          ]),
          e('h1', { cls: 'title is-2 title-gradient' }, 'Lazy Speech To Text'),
          e('p', { cls: 'subtitle has-text-grey' }, 'Convert your voice to text effortlessly')
        ]),
        e(LoginForm, { onLogin: handleLogin })
      ])
    ]);
  }

  // Show main app if authenticated
  return e('div', { style: { minHeight: '100vh', display: 'flex', flexDirection: 'column' } }, [
    e(Navbar, { username: authState.username, onLogout: handleLogout }),
    e('section', { cls: 'section', style: { paddingTop: '5rem', flex: '1' } }, [
      e('div', { cls: 'container' }, [
        e('div', { cls: 'card-elevated', style: { padding: '2rem', background: 'white' } }, [
          e(AppContent)
        ])
      ])
    ]),
    e(Footer)
  ]);
}

// Footer component
function Footer() {
  const currentYear = new Date().getFullYear();
  return e('footer', { 
    cls: 'footer', 
    style: { 
      padding: '1.5rem',
      background: 'linear-gradient(135deg, #0891b2 0%, #0d9488 100%)',
      color: 'white'
    } 
  }, [
    e('div', { cls: 'content has-text-centered' }, [
      e('p', { style: { marginBottom: '0.5rem' } }, [
        e('span', null, `© ${currentYear} Lazy Speech To Text Converter. All rights reserved.`)
      ]),
      e('p', { style: { fontSize: '0.9rem', opacity: '0.9' } }, [
        e('span', null, 'Created by '),
        e('a', { 
          href: 'mailto:walter.fan@gmail.com', 
          style: { color: 'white', textDecoration: 'underline' }
        }, 'walter.fan@gmail.com')
      ])
    ])
  ]);
}

document.addEventListener('DOMContentLoaded', () => {
  ReactDOM.render(e(App), document.getElementById('app'));
});