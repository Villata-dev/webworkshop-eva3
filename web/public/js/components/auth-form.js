// web/public/js/components/auth-form.js

// Define la plantilla HTML para el componente
// Usamos un <template> para definir la estructura que se clonará
const authFormTemplate = document.createElement('template');
authFormTemplate.innerHTML = `
    <style>
        /* Estilos encapsulados para el Shadow DOM */
        .auth-container {
            max-width: 400px;
            margin: 20px auto;
            padding: 20px;
            border: 1px solid #ccc;
            border-radius: 5px;
            box-shadow: 2px 2px 10px rgba(0, 0, 0, 0.1);
            text-align: center; /* Centrar contenido del formulario */
        }
        h2 {
            text-align: center;
            margin-bottom: 20px;
            color: #333;
        }
        .form-group {
            margin-bottom: 15px;
            text-align: left; /* Alinear etiquetas a la izquierda */
        }
        label {
            display: block; /* Cada etiqueta en su propia línea */
            margin-bottom: 5px;
            font-weight: bold;
            color: #555;
        }
        input[type="text"],
        input[type="password"] {
            width: calc(100% - 22px); /* Ancho completo menos padding y borde */
            padding: 10px;
            border: 1px solid #ccc;
            border-radius: 4px;
            box-sizing: border-box; /* Incluir padding y borde en el ancho */
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 10px 15px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            width: 100%; /* Botón al ancho completo */
            margin-top: 10px;
            transition: background-color 0.3s ease; /* Transición suave */
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

    </style>

    <div class="auth-container">
        <h2 id="form-title">Iniciar Sesión</h2>
        <div id="auth-form">
             </div>
        <p id="error-display" class="error-message" style="display: none;"></p>
        <p class="toggle-mode">
            <span id="toggle-text">¿No tienes cuenta?</span> <a href="#" id="toggle-link">Regístrate aquí</a>.
        </p>
    </div>
`;

// Define la clase Custom Element
class AuthForm extends HTMLElement {
    constructor() {
        super(); // Llama al constructor de HTMLElement

        // Adjunta el Shadow DOM al componente. 'open' permite acceso desde JS externo.
        this.attachShadow({ mode: 'open' });

        // Clona el contenido de la plantilla y adjúntalo al Shadow DOM
        this.shadowRoot.appendChild(authFormTemplate.content.cloneNode(true));

        // Referencias a elementos clave dentro del Shadow DOM
        this.$formTitle = this.shadowRoot.getElementById('form-title');
        this.$authForm = this.shadowRoot.getElementById('auth-form');
        this.$toggleText = this.shadowRoot.getElementById('toggle-text');
        this.$toggleLink = this.shadowRoot.getElementById('toggle-link');
        this.$errorDisplay = this.shadowRoot.getElementById('error-display');

        // Estado interno para saber si estamos en modo login o registro
        this._mode = 'login'; // Puede ser 'login' o 'register'

        // Añadir campos iniciales (vacíos por ahora) y lógica de toggle
        this._renderForm(); // Renderiza el formulario inicial (login)
        this._attachEventListeners(); // Añadir listeners a los botones, etc.

        console.log("AuthForm constructor ejecutado.");
    }

    // --- Ciclo de Vida del Componente ---

    // connectedCallback: Se llama cuando el elemento es adjuntado al DOM
    connectedCallback() {
        console.log("AuthForm añadido al DOM.");
        // Aquí podrías hacer setup inicial si es necesario
    }

    // disconnectedCallback: Se llama cuando el elemento es removido del DOM
    disconnectedCallback() {
        console.log("AuthForm removido del DOM.");
        // Aquí podrías hacer limpieza (remover listeners, etc.)
    }

    // attributeChangedCallback: Se llama cuando un atributo observado cambia, se añade o remueve
    // static get observedAttributes() { return ['mode']; } // Observar el atributo 'mode'
    // attributeChangedCallback(name, oldValue, newValue) {
    //     if (name === 'mode' && oldValue !== newValue) {
    //         this._mode = newValue;
    //         this._renderForm(); // Volver a renderizar si cambia el modo por atributo
    //     }
    // }


    // --- Métodos Internos ---

    _renderForm() {
        // Limpiar contenido previo del formulario
        this.$authForm.innerHTML = '';

        let usernameField = `
             <div class="form-group">
                <label for="username">Nombre de Usuario:</label>
                <input type="text" id="username" name="username" required>
            </div>
        `;
        let passwordField = `
             <div class="form-group">
                <label for="password">Contraseña:</label>
                <input type="password" id="password" name="password" required>
            </div>
        `;
        let submitButton;

        if (this._mode === 'login') {
            this.$formTitle.textContent = 'Iniciar Sesión';
            submitButton = '<button id="submit-btn">Iniciar Sesión</button>';
            this.$toggleText.textContent = '¿No tienes cuenta?';
            this.$toggleLink.textContent = 'Regístrate aquí';
        } else { // register mode
            this.$formTitle.textContent = 'Registrarse';
            submitButton = '<button id="submit-btn">Registrarse</button>';
             this.$toggleText.textContent = '¿Ya tienes cuenta?';
            this.$toggleLink.textContent = 'Inicia sesión aquí';
        }

        // Construir el formulario e inyectarlo
        this.$authForm.innerHTML = usernameField + passwordField + submitButton;
         this.hideError(); // Ocultar error al cambiar de modo
    }

    _attachEventListeners() {
        // Añadir listener al enlace de toggle mode
        this.$toggleLink.addEventListener('click', (e) => {
            e.preventDefault(); // Prevenir navegación por defecto
            this.toggleMode(); // Cambiar el modo del formulario
        });

        // Añadir listener al botón de submit (delegación de eventos)
        // Como el botón se re-renderiza, usamos un listener en el contenedor del formulario
        this.$authForm.addEventListener('click', (e) => {
            if (e.target && e.target.id === 'submit-btn') {
                this._handleSubmit(e); // Manejar el envío del formulario
            }
        });
    }

    _handleSubmit(event) {
         event.preventDefault(); // Prevenir envío del formulario por defecto

         this.hideError(); // Ocultar errores previos

         // Obtener valores de los campos
         const usernameInput = this.shadowRoot.getElementById('username');
         const passwordInput = this.shadowRoot.getElementById('password');

         const username = usernameInput.value.trim();
         const password = passwordInput.value.trim();

         if (!username || !password) {
             this.showError("Nombre de usuario y contraseña no pueden estar vacíos.");
             return; // Detener si los campos están vacíos
         }

         console.log(`Intentando ${this._mode} con usuario: ${username}`);

         // Aquí es donde se hará la llamada a la API (fetch)
         // Por ahora, solo logeamos la acción.
         // Después, implementaremos la función para llamar a la API
         if (this._mode === 'login') {
             this._callLoginApi(username, password);
         } else { // register
             this._callRegisterApi(username, password);
         }

         // Limpiar campos después de intentar enviar (opcional, pero buena UX)
         // usernameInput.value = '';
         // passwordInput.value = '';
    }

    // --- Métodos Públicos ---

    // Método para cambiar el modo del formulario (login/register)
    toggleMode() {
        this._mode = this._mode === 'login' ? 'register' : 'login';
        this._renderForm(); // Volver a renderizar el formulario para el nuevo modo
        console.log(`Modo de formulario cambiado a: ${this._mode}`);
    }

    // Método para mostrar un mensaje de error
    showError(message) {
        this.$errorDisplay.textContent = message;
        this.$errorDisplay.style.display = 'block';
    }

     // Método para ocultar el mensaje de error
    hideError() {
        this.$errorDisplay.textContent = '';
        this.$errorDisplay.style.display = 'none';
    }

    // Placeholder para la llamada a la API de Login (la implementaremos pronto)
    _callLoginApi(username, password) {
         console.log(`Llamando a la API de Login para ${username}...`);
         // TODO: Implementar fetch a /api/auth/login
         this.showError("Login aún no implementado. Ver consola."); // Mensaje temporal
    }

    // Placeholder para la llamada a la API de Registro (la implementaremos pronto)
    _callRegisterApi(username, password) {
        console.log(`Llamando a la API de Registro para ${username}...`);
        // TODO: Implementar fetch a /api/auth/register
         this.showError("Registro aún no implementado. Ver consola."); // Mensaje temporal
    }

    // Podemos añadir métodos públicos para, por ejemplo, establecer el modo desde fuera
    // set mode(value) {
    //    if (value === 'login' || value === 'register') {
    //         this._mode = value;
    //         this._renderForm();
    //     } else {
    //         console.error("Modo inválido para AuthForm. Debe ser 'login' o 'register'.");
    //     }
    // }
    // get mode() { return this._mode; }

}

// 4. Definir el Custom Element en el registro del navegador
// El nombre del tag DEBE contener un guion (-)
customElements.define('auth-form', AuthForm);

console.log("Componente auth-form.js cargado y definido.");