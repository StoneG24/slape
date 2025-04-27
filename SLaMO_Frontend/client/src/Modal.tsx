import ReactNode from "react";
import "./modal.css";

interface modalProperties {
  isOpen: boolean;
  onClose: () => void;
  children: ReactNode;
}

export default function Modal({isOpen, onClose, children}: modalProperties) {
  if (!isOpen) {
    return null;
  }

  if (localStorage.getItem("PromptSetting") == null)
    localStorage.setItem("PromptSetting", "Automatic");
  if (localStorage.getItem("StyleSetting") == null)
    localStorage.setItem("StyleSetting", "Dark");

  const themeColor: string | null = localStorage.getItem("StyleSetting");

  return (
    <>
      <div className={`${themeColor}_overlay`}>
        <div className="modal-content">
          <button className={`${themeColor}_modal-close`} onClick={onClose}>
            &times;
          </button>
          {children}
        </div>
      </div>
      <div className="dimGuy"></div>
    </>
  );
}
