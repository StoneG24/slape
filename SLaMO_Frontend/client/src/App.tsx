import { BrowserRouter, Routes, Route } from "react-router-dom";
import Prompt from "./Prompt.tsx";
import Settings from "./Settings.tsx";
import Pipelines from "./Pipelines.tsx";
import Logs from "./Logs.tsx";

// I'm envisioning this as being the function/component that just houses the currently selected page.
export default function App() {

    // You'd have the logic for which page is currently selected. Still not really sure how this will work but we'll figure it out
    return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Prompt/>}/>
        <Route path="/settings" element={<Settings/>}/>
        <Route path="/pipelines" element={<Pipelines/>}/>
        <Route path="/logs" element={<Logs/>}/>
      </Routes>
    </BrowserRouter>
    );
}

// Stone, we may need to brainstorm how we can change pages here. I'm trying to think, but I can't seem to land on a good way to do it. Maybe we can pass a pointer to a string or something that selects which page is currently active?