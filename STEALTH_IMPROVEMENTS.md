# Stealth Mode Improvements

## Overview
Updated the typing indicator implementation to make sessions appear as real human users rather than automated bots.

## Changes Made

### 1. **Realistic Push Names** ✅
**Before:**
- Default push name: "WhatsApp Bot" 
- Obvious automation indicator

**After:**
- Random realistic names from a pool of gender-neutral names
- Examples: "Alex", "Sam", "Jordan", "Taylor", "Casey", "Morgan", etc.
- Falls back to session name if provided by user

**Implementation:**
```go
func (s *WhatsAppService) generateRandomName() string {
    firstNames := []string{
        "Alex", "Sam", "Jordan", "Taylor", "Casey", "Morgan", "Jamie", "Riley",
        "Avery", "Peyton", "Quinn", "Sage", "Rowan", "Emery", "Hayden", "Finley",
        // ... more realistic names
    }
    // Crypto-secure random selection
}
```

### 2. **Natural Naming Strategy**
- **Gender-neutral names**: Avoids assumptions and seems more natural
- **Common modern names**: Names that are popular and don't raise suspicion
- **Single names only**: Many WhatsApp users use just first names
- **No "Bot" or automation keywords**: Completely removed bot-related terms

### 3. **User-Controlled Names**
Users can set their own realistic names when creating sessions:
```json
{
  "name": "Alex Johnson",
  "phone": "628123456789",
  "webhook_url": "https://example.com/webhook"
}
```

If no name is provided, a random realistic name is automatically assigned.

### 4. **Improved Documentation**
- Updated all examples to use realistic human names
- Removed references to "bot" or "automated" in user-facing documentation
- API examples now show natural naming patterns

## Stealth Best Practices

### For Users:
1. **Use realistic names**: Choose common names that don't stand out
2. **Match local culture**: Use names appropriate for your region
3. **Keep it simple**: Single first names work best
4. **Avoid patterns**: Don't use sequential names like "User1", "User2"

### For Developers:
1. **Random name pool**: Currently 38 diverse names in rotation
2. **Crypto-secure randomization**: Uses `crypto/rand` for unpredictability
3. **No telltale patterns**: Each session gets a different random name
4. **Fallback handling**: Graceful degradation if randomization fails

## Technical Implementation

### Name Generation Algorithm:
1. Check if user provided a name in session creation
2. If name exists, use it as push name
3. If no name, generate random name from pool
4. Use crypto/rand for secure randomization
5. Fallback to timestamp-based selection if crypto fails

### Security Considerations:
- **No correlation**: Random names don't correlate to session IDs
- **Non-sequential**: No predictable patterns in name assignment
- **Diverse pool**: Large enough pool to avoid frequent repeats
- **Cultural neutrality**: Names work across different regions

## Impact on Typing Indicators

The realistic push names ensure:
- ✅ Typing indicators work properly (technical requirement)
- ✅ Recipients see normal human names, not "Bot" or automation hints
- ✅ Sessions blend in with regular WhatsApp users
- ✅ No obvious signs of automation to casual observers

## Future Enhancements

Potential improvements for even better stealth:
1. **Regional name pools**: Different names based on phone number country code
2. **Full name support**: First + Last name combinations
3. **Profile picture randomization**: Random avatars (if supported by whatsmeow)
4. **Status message variety**: Random status messages like real users
5. **Activity patterns**: Mimic human usage patterns for presence

## Usage Examples

### Automatic Random Name:
```bash
curl -X POST http://localhost:3000/api/sessions \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"webhook_url": ""}'
# System assigns random name like "Riley" or "Cameron"
```

### User-Specified Name:
```bash
curl -X POST http://localhost:3000/api/sessions \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Sarah Chen", "webhook_url": ""}'
# Uses "Sarah Chen" as push name
```

This approach ensures maximum stealth while maintaining full functionality of typing indicators and presence features.