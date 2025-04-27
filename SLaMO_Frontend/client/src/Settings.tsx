import DropDownButton from "./DropDownButton.tsx";
import MenuTabs from "./MenuTabs.tsx";
import {useState} from "react";
import "./settings.css";
import "./index.css";

export default function Settings() {
  if (localStorage.getItem("PromptSetting") == null)
    localStorage.setItem("PromptSetting", "Automatic");
  if (localStorage.getItem("StyleSetting") == null)
    localStorage.setItem("StyleSetting", "Dark");

  const themeColor: string | null = localStorage.getItem("StyleSetting");

  const [PromptSetting, setPromptSetting] = useState(
    localStorage.getItem("PromptSetting")
  );
  const [StyleSetting, setStyleSetting] = useState(
    localStorage.getItem("StyleSetting")
  );

  const promptDropDownOptions = [
    {type: "Automatic", name: "Automatic"},
    {type: "Manual", name: "Manual"},
    {type: "Mixed", name: "Mixed"},
  ];
  const styleDropDownOptions = [
    {type: "Pink", name: "Pink"},
    {type: "Dark", name: "Dark"},
    {type: "Light", name: "Light"},
  ];

  function settingsButtonHandler() {
    localStorage.setItem("PromptSetting", PromptSetting);
    localStorage.setItem("StyleSetting", StyleSetting);
    globalThis.location.reload();
  }

  return (
    <>
      <div className={`${themeColor}_background`} />
      <MenuTabs />
      <div className={`${themeColor}_settingsDiv`}>
        <h1 className={`${themeColor}_settingTitle`}>Settings</h1>
        <h2>Theme</h2>
        <DropDownButton
          className="colorSetting"
          value={StyleSetting}
          callBack={setStyleSetting}
          optionObject={styleDropDownOptions}
        />
        <hr />
        <button
          className={`${themeColor}_Submit`}
          onClick={settingsButtonHandler}
        >
          Display Settings
        </button>
      </div>
    </>
  );
}
