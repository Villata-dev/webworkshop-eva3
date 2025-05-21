// web/public/js/components/auth-form.js

// Define la plantilla HTML para el componente
const authFormTemplate = document.createElement('template');
authFormTemplate.innerHTML = `
    <style>
        /* Mantén los estilos que ya añadiste */
        .auth-container {
            max-width: 400px;
            margin: 20px auto;
            padding: 20px;
            border: 1px solid #ccc;
            border-radius: 5px;
            box-shadow: 2px 2px 10px rgba(0, 0, 0, 0.1);
            text-align: center;
        }
        h2 {
            text-align: center;
            margin-bottom: 20px;
            color: #333;
        }
        .form-group {
            margin-bottom: 15px;
            text-align: left;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #555;
        }
        input[type="text"],
        input[type="password"] {
            width: calc(100% - 22px);
            padding: 10px;
            border: 1px solid #ccc;
            border-radius: 4px;
            box-sizing: border-box;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 10px 15px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            width: 100%;
            margin-top: 10px;
            transition: background-color 0.3s ease;
        }
        button:hover {
            background-color: #0056b3;
        }
        .toggle-mode {
            margin-top: 15px;
            font-size: 0.9em;
        }
        .toggle-mode a {
            color: #007bff;
            text-decoration: none;
            cursor: pointer;
        }
        .toggle-mode a:hover {
             text-decoration: underline;
        }
         .error-message {
            color: red;
            font-size: 0.9em;
            margin-top: 10px;
            text-align: center;
        }

        .auth-form button {
            opacity: 1;
            transition: opacity 0.3s ease;
        }

        .auth-form button:disabled {
            opacity: 0.6;
            cursor: not-allowed;
        }

        .success-message {
            color: green;
            font-size: 0.9em;
            margin-top: 10px;
            text-align: center;
            opacity: 0;
            transition: opacity 0.3s ease;
        }

        .success-message.visible {
            opacity: 1;
        }

        .loading {
            position: relative;
        }

        .loading:after {
            content: '';
            position: absolute;
            width: 20px;
            height: 20px;
            border: 2px solid #f3f3f3;
            border-top: 2px solid #3498db;
            border-radius: 50%;
            right: 10px;
            top: 50%;
            transform: translateY(-50%);
            animation: spin 1s linear infinite;
        }

        @keyframes spin {
            0% { transform: translateY(-50%) rotate(0deg); }
            100% { transform: translateY(-50%) rotate(360deg); }
        }

    </style>

    <div class="auth-container">
        <h2 id="form-title">Iniciar Sesión</h2>
        <div id="auth-form">
             </div>
        <p id="error-display" class="error-message" style="display: none;"></p>
        <p class="toggle-mode">
            <span id="toggle-text"></span> <a href="#" id="toggle-link"></a>.
        </p>
    </div>
`;

// Define la clase Custom Element
class AuthForm extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        console.log('AuthForm constructor ejecutado.');
        this._renderForm();
    }

    _renderForm() {
        this.shadowRoot.innerHTML = `
            <style>
                .login-form {
                    max-width: 300px;
                    margin: 20px auto;
                    padding: 20px;
                    border: 1px solid #ddd;
                    border-radius: 8px;
                    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                }
                input {
                    width: 100%;
                    padding: 8px;
                    margin: 8px 0;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }
                button {
                    width: 100%;
                    padding: 10px;
                    background-color: #007bff;
                    color: white;
                    border: none;
                    border-radius: 4px;
                    cursor: pointer;
                }
                button:hover {
                    background-color: #0056b3;
                }
                .error-message {
                    color: red;
                    margin-top: 10px;
                    text-align: center;
                }
            </style>
            <div class="login-form">
                <form id="loginForm">
                    <h2>Iniciar Sesión</h2>
                    <div>
                        <label for="username">Usuario:</label>
                        <input type="text" id="username" required>
                    </div>
                    <div>
                        <label for="password">Contraseña:</label>
                        <input type="password" id="password" required>
                    </div>
                    <button type="submit">Iniciar Sesión</button>
                    <div class="error-message"></div>
                </form>
            </div>
        `;

        this.shadowRoot.querySelector('#loginForm').addEventListener('submit', this._handleSubmit.bind(this));
    }

    async _handleSubmit(e) {
        e.preventDefault();
        console.log('Intentando iniciar sesión...');

        const username = this.shadowRoot.getElementById('username').value;
        const password = this.shadowRoot.getElementById('password').value;
        const errorMessage = this.shadowRoot.querySelector('.error-message');

        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: password
                }),
                credentials: 'include'
            });

            console.log('Respuesta del servidor:', response.status);

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Error en el inicio de sesión');
            }

            const userData = await response.json();
            console.log('Login exitoso:', userData);

            this.dispatchEvent(new CustomEvent('auth-success', {
                bubbles: true,
                composed: true,
                detail: { user: userData }
            }));

            errorMessage.textContent = '';
        } catch (error) {
            console.error('Error en login:', error);
            errorMessage.textContent = error.message || 'Error en el inicio de sesión';
        }
    }
}

customElements.define('auth-form', AuthForm);
console.log('Componente auth-form registrado');