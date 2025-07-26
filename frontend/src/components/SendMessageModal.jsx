import { useState } from 'react';
import axios from 'axios';

const SendMessageModal = ({ session, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    to: '',
    message: '',
  });
  const [sending, setSending] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.to || !formData.message) return;

    setSending(true);
    
    try {
      const response = await axios.post(`/api/sessions/${session.id}/send`, {
        to: formData.to,
        message: formData.message,
      });

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

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold">Send Message</h3>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700"
          >
            <i className="fas fa-times text-xl"></i>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              To:
            </label>
            <input
              type="text"
              name="to"
              value={formData.to}
              onChange={handleChange}
              placeholder="6281234567890 (or with @s.whatsapp.net)"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
              required
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Message:
            </label>
            <textarea
              name="message"
              value={formData.message}
              onChange={handleChange}
              placeholder="Type your message here..."
              rows="4"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
              required
            />
          </div>
          
          <div className="flex gap-3">
            <button
              type="submit"
              disabled={sending || !formData.to || !formData.message}
              className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white px-4 py-2 rounded-lg font-medium transition duration-200"
            >
              {sending ? (
                <>
                  <i className="fas fa-spinner fa-spin mr-2"></i>
                  Sending...
                </>
              ) : (
                <>
                  <i className="fas fa-paper-plane mr-2"></i>
                  Send Message
                </>
              )}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition duration-200"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default SendMessageModal;