import { useState } from 'react';
import axios from 'axios';

const SendMessageModal = ({ session, onClose, onSuccess }) => {
  const [messageType, setMessageType] = useState('text'); // 'text' or 'location'
  const [formData, setFormData] = useState({
    to: '',
    message: '',
    latitude: '',
    longitude: '',
  });
  const [sending, setSending] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.to) return;
    
    if (messageType === 'text' && !formData.message) return;
    if (messageType === 'location' && (!formData.latitude || !formData.longitude)) return;

    setSending(true);
    
    try {
      let response;
      
      if (messageType === 'text') {
        response = await axios.post(`/api/sessions/${session.id}/send`, {
          to: formData.to,
          message: formData.message,
        });
      } else if (messageType === 'location') {
        response = await axios.post(`/api/sessions/${session.id}/send-location`, {
          to: formData.to,
          latitude: parseFloat(formData.latitude),
          longitude: parseFloat(formData.longitude),
        });
      }

      if (response.data.success) {
        onSuccess();
        onClose();
      }
    } catch (error) {
      console.error('Error sending message:', error);
    } finally {
      setSending(false);
    }
  };

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const getCurrentLocation = () => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          setFormData({
            ...formData,
            latitude: position.coords.latitude.toString(),
            longitude: position.coords.longitude.toString(),
          });
        },
        (error) => {
          console.error('Error getting location:', error);
          alert('Unable to retrieve your location. Please enter coordinates manually.');
        }
      );
    } else {
      alert('Geolocation is not supported by this browser.');
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content max-w-lg w-full animate-scale-in">
        <div className="card-header">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-primary-500 to-primary-600 rounded-xl flex items-center justify-center shadow-sm">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                </svg>
              </div>
              <div>
                <h3 className="text-title">Send Message</h3>
                <p className="text-caption">Session: {session?.name || `#${session?.id}`}</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="btn btn-ghost p-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <div className="card-body">
          {/* Message Type Selector */}
          <div className="mb-6">
            <div className="text-overline mb-3">Message Type</div>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setMessageType('text')}
                className={`btn btn-sm flex-1 ${messageType === 'text' ? 'btn-primary' : 'btn-secondary'}`}
              >
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                Text Message
              </button>
              <button
                type="button"
                onClick={() => setMessageType('location')}
                className={`btn btn-sm flex-1 ${messageType === 'location' ? 'btn-primary' : 'btn-secondary'}`}
              >
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                Send Location
              </button>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Recipient */}
            <div className="input-group">
              <label className="input-label">Recipient</label>
              <input
                type="text"
                name="to"
                value={formData.to}
                onChange={handleChange}
                placeholder="6281234567890 or 6281234567890@s.whatsapp.net"
                className="input"
                required
              />
              <div className="input-helper">Enter WhatsApp number with country code</div>
            </div>

            {/* Text Message Fields */}
            {messageType === 'text' && (
              <div className="input-group">
                <label className="input-label">Message</label>
                <textarea
                  name="message"
                  value={formData.message}
                  onChange={handleChange}
                  placeholder="Type your message here..."
                  rows="4"
                  className="input"
                  required
                />
                <div className="input-helper">Enter the text message to send</div>
              </div>
            )}

            {/* Location Fields */}
            {messageType === 'location' && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="text-body-medium">Location Coordinates</div>
                  <button
                    type="button"
                    onClick={getCurrentLocation}
                    className="btn btn-sm btn-secondary"
                  >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Use Current Location
                  </button>
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="input-group">
                    <label className="input-label">Latitude</label>
                    <input
                      type="number"
                      name="latitude"
                      value={formData.latitude}
                      onChange={handleChange}
                      placeholder="-6.200000"
                      step="any"
                      className="input"
                      required
                    />
                  </div>
                  <div className="input-group">
                    <label className="input-label">Longitude</label>
                    <input
                      type="number"
                      name="longitude"
                      value={formData.longitude}
                      onChange={handleChange}
                      placeholder="106.816666"
                      step="any"
                      className="input"
                      required
                    />
                  </div>
                </div>
                
                {formData.latitude && formData.longitude && (
                  <div className="p-4 bg-gray-50 rounded-lg border border-gray-200">
                    <div className="text-body-medium mb-2">Preview Location</div>
                    <div className="text-caption text-gray-600">
                      Latitude: {formData.latitude}<br />
                      Longitude: {formData.longitude}
                    </div>
                    <a
                      href={`https://www.google.com/maps?q=${formData.latitude},${formData.longitude}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary-600 hover:text-primary-700 text-caption underline mt-2 inline-block"
                    >
                      View on Google Maps →
                    </a>
                  </div>
                )}
              </div>
            )}
          </form>
        </div>

        <div className="card-footer">
          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="btn btn-secondary flex-1"
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={sending || !formData.to || 
                (messageType === 'text' && !formData.message) ||
                (messageType === 'location' && (!formData.latitude || !formData.longitude))
              }
              className="btn btn-primary flex-1"
            >
              {sending ? (
                <>
                  <div className="loading-spinner mr-2"></div>
                  Sending...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                  </svg>
                  Send {messageType === 'text' ? 'Message' : 'Location'}
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SendMessageModal;