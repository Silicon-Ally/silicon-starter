export interface PublicRuntimeConfig {
  firebaseConfig: FirebaseConfig,
}
  
export interface FirebaseConfig {
  apiKey: string,
  appId: string,
  authDomain: string,
  projectId: string,
  storageBucket: string,
  messagingSenderId: string,
  vapidKey: string,
  serviceWorker: string,
}
