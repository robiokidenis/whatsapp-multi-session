import { useState, useEffect, useRef } from 'react';
import QRCode from 'qrcode';
import { useAuth } from '../contexts/AuthContext';

const QRModal = ({ session, onClose, onSuccess }) => {
  const { token } = useAuth();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);
  const qrElementRef = useRef(null);
  const wsRef = useRef(null);
  const countdownIntervalRef = useRef(null);

  useEffect(() => {
    // Prevent double connection in React strict mode
    if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
      return;
    }
    
    connectWebSocket();
    
    return () => {
      cleanup();
    };
  }, [session.id, token]); // Add dependencies to prevent unnecessary reconnections

  const cleanup = () => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (countdownIntervalRef.current) {
      clearInterval(countdownIntervalRef.current);
      countdownIntervalRef.current = null;
    }
  };

  const connectWebSocket = () => {
    // Prevent multiple connections
    if (wsRef.current && wsRef.current.readyState === WebSocket.CONNECTING) {
      console.log('WebSocket already connecting, skipping...');
      return;
    }
    
    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Use Vite proxy for WebSocket connection
    const wsUrl = `${wsProtocol}//${window.location.host}/api/ws/${session.id}?token=${encodeURIComponent(token)}`;
    
    console.log('Connecting to WebSocket:', wsUrl);
    wsRef.current = new WebSocket(wsUrl);
    
    wsRef.current.onopen = () => {
      console.log('WebSocket connection opened');
    };
    
    wsRef.current.onmessage = (event) => {
      console.log('🔥 WebSocket message received:', event.data);
      try {
        const data = JSON.parse(event.data);
        console.log('📋 Parsed data:', data);
        
        if (data.type === 'qr') {
          console.log('🎯 QR message detected, generating QR code...');
          setLoading(false);
          generateQRCode(data.data.qr);
          
          // Start countdown
          const timeoutSeconds = Math.floor(data.data.timeout / 1000000000);
          console.log('⏰ Setting countdown to:', timeoutSeconds, 'seconds');
          setCountdown(timeoutSeconds);
          startCountdown();
        } else if (data.type === 'success') {
          console.log('✅ Success message received');
          setLoading(false);
          onSuccess();
          onClose();
        } else if (data.type === 'error') {
          console.log('❌ Error message received:', data.error);
          setLoading(false);
          setError(data.error);
        } else {
          console.log('❓ Unknown message type:', data.type);
        }
      } catch (parseError) {
        console.error('Failed to parse WebSocket message:', parseError);
        console.log('Raw message:', event.data);
      }
    };
    
    wsRef.current.onerror = (error) => {
      console.error('WebSocket error:', error);
      setLoading(false);
      setError('Failed to connect to server. Please check if the backend is running on port 8080.');
    };
    
    wsRef.current.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      if (event.code === 409) {
        setError('QR generation already in progress. Please wait and try again.');
      } else if (loading) {
        setLoading(false);
        setError(`Connection closed (${event.code}): ${event.reason || 'Unknown error'}`);
      }
    };
  };

  const generateQRCode = async (qrString) => {
    console.log('SIMPLE QR: Got data:', qrString);
    
    // Wait for DOM to render, then try multiple ways to find element
    const findAndRenderQR = async () => {
      let qrElement = qrElementRef.current;
      if (!qrElement) {
        console.log('SIMPLE QR: Ref null, trying getElementById...');
        qrElement = document.getElementById('qr-container');
      }
      if (!qrElement) {
        console.log('SIMPLE QR: getElementById failed, trying querySelector...');
        qrElement = document.querySelector('.qr-container');
      }
      if (!qrElement) {
        console.log('SIMPLE QR: querySelector failed, trying any div...');
        qrElement = document.querySelector('#root div[style*="dashed"]');
      }
      
      if (!qrElement) {
        console.log('SIMPLE QR: Still no element found, checking DOM...');
        console.log('Available elements:', document.querySelectorAll('div').length);
        return false;
      }
      
      console.log('SIMPLE QR: Found element, generating QR with package...');
      
      // Clear the element first
      qrElement.innerHTML = '';
      
      try {
        // Use QRCode package for local generation
        const canvas = document.createElement('canvas');
        await QRCode.toCanvas(canvas, qrString, {
          width: 300,
          margin: 2,
          color: {
            dark: '#000000',
            light: '#FFFFFF'
          }
        });
        
        // Style the canvas
        canvas.style.cssText = `
          max-width: 100%; 
          height: auto; 
          border: 1px solid #ccc; 
          border-radius: 8px; 
          background: white;
          display: block;
          margin: 0 auto;
        `;
        
        // Add container with canvas
        const container = document.createElement('div');
        container.style.cssText = 'text-align: center; padding: 20px;';
        container.appendChild(canvas);
        qrElement.appendChild(container);
        
        console.log('QR CODE GENERATED WITH PACKAGE');
        
      } catch (error) {
        console.error('QRCode package failed:', error);
        
        // Fallback to external API only if package fails
        qrElement.innerHTML = `
          <div style="text-align: center; padding: 20px;">
            <img 
              src="https://api.qrserver.com/v1/create-qr-code/?size=256x256&data=${encodeURIComponent(qrString)}" 
              alt="QR Code"
              style="max-width: 100%; border: 1px solid #ccc; background: white;"
              onload="console.log('QR FALLBACK API LOADED')"
              onerror="console.log('QR FALLBACK API FAILED');"
            />
          </div>
        `;
      }
      return true;
    };
    
    // Try immediately first
    if (!(await findAndRenderQR())) {
      console.log('SIMPLE QR: Immediate attempt failed, waiting 100ms...');
      setTimeout(async () => {
        if (!(await findAndRenderQR())) {
          console.log('SIMPLE QR: Still failed, waiting 500ms...');
          setTimeout(async () => {
            if (!(await findAndRenderQR())) {
              console.log('SIMPLE QR: Final attempt failed. Modal might not be open.');
            }
          }, 500);
        }
      }, 100);
    }
  };


  const startCountdown = () => {
    countdownIntervalRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(countdownIntervalRef.current);
          setError('QR code expired');
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  const handleClose = () => {
    cleanup();
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold">Scan QR Code</h3>
          <button
            onClick={handleClose}
            className="text-gray-500 hover:text-gray-700"
          >
            <i className="fas fa-times text-xl"></i>
          </button>
        </div>

        <div className="text-center">
          {loading && (
            <div className="py-8">
              <i className="fas fa-spinner fa-spin text-4xl text-gray-400 mb-4"></i>
              <p className="text-gray-600">Generating QR Code...</p>
            </div>
          )}

          {!loading && !error && (
            <div className="space-y-4">
              <div 
                ref={qrElementRef} 
                id="qr-container" 
                className="flex justify-center qr-container"
                style={{minHeight: '200px', border: '2px dashed #ccc', padding: '10px'}}
              >
                <div style={{color: '#666', fontStyle: 'italic'}}>QR code will appear here...</div>
              </div>
              <p className="text-sm text-gray-600">
                Scan this QR code with WhatsApp on your phone
              </p>
              {countdown > 0 && (
                <div className="text-sm text-orange-600">
                  Expires in {countdown} seconds
                </div>
              )}
            </div>
          )}

          {error && (
            <div className="py-8 text-red-600">
              <i className="fas fa-exclamation-circle text-4xl mb-4"></i>
              <p>{error}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default QRModal;