class AIAssistant {
    constructor() {
        this.conversation = [];
        this.currentModel = 'llama3.2';
        this.isGenerating = false;
        
        this.init();
    }
    
    init() {
        this.setupEventListeners();
        this.loadModels();
        this.checkHealth();
        
        // Check health every 30 seconds
        setInterval(() => this.checkHealth(), 30000);
        
        // Welcome message
        this.addMessage('system', 'Welcome to AI Assistant! Select a model and start chatting.');
    }
    
    setupEventListeners() {
        const sendButton = document.getElementById('send-button');
        const messageInput = document.getElementById('message-input');
        const newChatButton = document.getElementById('new-chat');
        const modelSelect = document.getElementById('model-select');
        const temperatureSlider = document.getElementById('temperature');
        const tempValue = document.getElementById('temp-value');
        
        sendButton.addEventListener('click', () => this.sendMessage());
        messageInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });
        
        newChatButton.addEventListener('click', () => this.newChat());
        modelSelect.addEventListener('change', (e) => {
            this.currentModel = e.target.value;
            this.addMessage('system', `Switched to model: ${e.target.value}`);
        });
        
        temperatureSlider.addEventListener('input', (e) => {
            tempValue.textContent = e.target.value;
        });
    }
    
    async loadModels() {
        try {
            const response = await fetch('/api/models');
            const data = await response.json();
            
            const modelSelect = document.getElementById('model-select');
            modelSelect.innerHTML = '';
            
            if (data.models && data.models.length > 0) {
                data.models.forEach(model => {
                    const option = document.createElement('option');
                    option.value = model.name;
                    option.textContent = `${model.name} (${model.size})`;
                    modelSelect.appendChild(option);
                });
                
                this.currentModel = data.models[0].name;
            } else {
                const option = document.createElement('option');
                option.value = 'llama3.2';
                option.textContent = 'llama3.2 (default)';
                modelSelect.appendChild(option);
            }
        } catch (error) {
            console.error('Failed to load models:', error);
            this.addMessage('system', 'Failed to load available models. Using default.');
        }
    }
    
    async checkHealth() {
        const statusElement = document.getElementById('connection-status');
        
        try {
            const response = await fetch('/api/health');
            const data = await response.json();
            
            if (data.status === 'healthy') {
                statusElement.textContent = '✅ Connected';
                statusElement.className = 'status-connected';
                
                // Check Ollama connection specifically
                if (data.connections && data.connections.ollama === 'connected') {
                    statusElement.textContent = '✅ AI Ready';
                } else {
                    statusElement.textContent = '⚠️ AI Disconnected';
                    statusElement.className = 'status-disconnected';
                }
            } else {
                statusElement.textContent = '❌ Server Error';
                statusElement.className = 'status-disconnected';
            }
        } catch (error) {
            statusElement.textContent = '❌ Offline';
            statusElement.className = 'status-disconnected';
        }
    }
    
    async sendMessage() {
        const messageInput = document.getElementById('message-input');
        const sendButton = document.getElementById('send-button');
        const message = messageInput.value.trim();
        
        if (!message || this.isGenerating) return;
        
        // Disable input while generating
        this.isGenerating = true;
        messageInput.value = '';
        sendButton.disabled = true;
        sendButton.textContent = 'Sending...';
        
        // Add user message to conversation
        this.addMessage('user', message);
        
        // Show typing indicator
        const typingId = this.addMessage('assistant', '', true);
        
        try {
            const streamMode = document.getElementById('stream-mode').checked;
            
            if (streamMode) {
                await this.sendStreamMessage(message, typingId);
            } else {
                await this.sendRegularMessage(message, typingId);
            }
        } catch (error) {
            console.error('Error sending message:', error);
            this.updateMessage(typingId, '❌ Error: Failed to get response from AI. Check if Ollama is running.');
        } finally {
            // Re-enable input
            this.isGenerating = false;
            sendButton.disabled = false;
            sendButton.textContent = 'Send';
            messageInput.focus();
        }
    }
    
    async sendRegularMessage(message, typingId) {
        const temperature = parseFloat(document.getElementById('temperature').value);
        
        const response = await fetch('/api/chat', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                model: this.currentModel,
                conversation_history: this.conversation.filter(msg => msg.role !== 'system'),
                temperature: temperature
            })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        // Update the typing indicator with actual response
        this.updateMessage(typingId, data.response);
        
        // Add to conversation history
        this.conversation.push({
            role: 'user',
            content: message,
            timestamp: new Date().toISOString()
        });
        this.conversation.push({
            role: 'assistant',
            content: data.response,
            timestamp: new Date().toISOString()
        });
    }
    
    async sendStreamMessage(message, typingId) {
        const temperature = parseFloat(document.getElementById('temperature').value);
        
        const response = await fetch('/api/chat/stream', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                model: this.currentModel,
                conversation_history: this.conversation.filter(msg => msg.role !== 'system'),
                temperature: temperature
            })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let fullResponse = '';
        
        // Remove typing indicator
        this.updateMessage(typingId, '');
        
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                
                const chunk = decoder.decode(value, { stream: true });
                const lines = chunk.split('\n');
                
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        
                        if (data === 'done') {
                            break;
                        } else if (data.startsWith('error:')) {
                            throw new Error(data.slice(6));
                        } else {
                            fullResponse += data;
                            this.updateMessage(typingId, fullResponse);
                        }
                    }
                }
            }
        } finally {
            reader.releaseLock();
        }
        
        // Add to conversation history
        this.conversation.push({
            role: 'user',
            content: message,
            timestamp: new Date().toISOString()
        });
        this.conversation.push({
            role: 'assistant',
            content: fullResponse,
            timestamp: new Date().toISOString()
        });
    }
    
    addMessage(role, content, isTyping = false) {
        const messagesContainer = document.getElementById('messages');
        const messageDiv = document.createElement('div');
        const messageId = 'msg-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
        
        messageDiv.id = messageId;
        messageDiv.className = `message ${role}`;
        
        if (isTyping) {
            messageDiv.innerHTML = '<div class="typing-indicator"></div>';
        } else {
            messageDiv.textContent = content;
        }
        
        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
        
        return messageId;
    }
    
    updateMessage(messageId, content) {
        const messageElement = document.getElementById(messageId);
        if (messageElement) {
            messageElement.textContent = content;
            
            // Auto-scroll to bottom
            const messagesContainer = document.getElementById('messages');
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
    }
    
    newChat() {
        // Clear conversation history
        this.conversation = [];
        
        // Clear messages display
        const messagesContainer = document.getElementById('messages');
        messagesContainer.innerHTML = '';
        
        // Add welcome message
        this.addMessage('system', 'New conversation started. How can I help you?');
        
        // Focus on input
        document.getElementById('message-input').focus();
    }
}

// Initialize the AI Assistant when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.aiAssistant = new AIAssistant();
});

// Handle page visibility changes (pause/resume health checks)
document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
        // Page is hidden, you could pause health checks here if needed
        console.log('Page hidden');
    } else {
        // Page is visible, ensure we have latest status
        console.log('Page visible');
        if (window.aiAssistant) {
            window.aiAssistant.checkHealth();
        }
    }
});