import MenuTabs from "./MenuTabs.tsx";
import logImageImportantDoNotRemove from "./logs.jpg";

export default function Logs() {

  function logHandler() {
    alert("Logger? I hardly know her");
  }

  if (localStorage.getItem("PromptSetting") == null)
    localStorage.setItem("PromptSetting", "Automatic");
  if (localStorage.getItem("StyleSetting") == null)
    localStorage.setItem("StyleSetting", "Dark");

  const themeColor: string | null = localStorage.getItem("StyleSetting");

  return (
    <>
      <div className={`${themeColor}_background`}/>
      <MenuTabs />
      <div>
        <img
          onClick={logHandler}
          src={logImageImportantDoNotRemove}
          alt="Image"
        />
      </div>
    </>
  );
}
