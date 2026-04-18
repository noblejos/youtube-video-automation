import { Routes, Route } from 'react-router-dom';
import { ToastProvider } from './ToastContext';
import Layout from './components/Layout';
import ToastContainer from './components/Toast';
import DashboardPage from './pages/DashboardPage';
import CreatePage from './pages/CreatePage';
import ProjectDetailPage from './pages/ProjectDetailPage';

function App() {
  return (
    <ToastProvider>
      <Layout>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/create" element={<CreatePage />} />
          <Route path="/projects/:id" element={<ProjectDetailPage />} />
        </Routes>
      </Layout>
      <ToastContainer />
    </ToastProvider>
  );
}

export default App;
