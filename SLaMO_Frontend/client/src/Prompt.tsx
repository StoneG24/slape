import {useState} from "react";
import "./App.css";
import MenuTabs from "./MenuTabs.tsx";
// import fetch from "node-fetch";
// Had to comment this out for my laptop :(
// import {colors} from "../../../../../../AppData/Local/deno/npm/registry.npmjs.org/debug/4.3.7/src/browser.js";

export default function Prompt() {
  const [PromptInfo, setPromptInfo] = useState(""); //used to contain the current value, and to set the new value

  // async function handleSubmit(event: {preventDefault: () => void}){
  //   event.preventDefault(); //makes sure the page doesn't reload when submitting the form
  //   const response = await fetch("localhost:3069/simple", {
  //     method:"POST",
  //     headers: {
  //       "Content-Type": "application/json"
  //     },
  //     body: JSON.stringify({prompt: prompt})
  //   });
  //   alert(response);
  //   setPromptInfo(""); //clears the prompt box after submission
  // };

  async function handleSubmit(event: {preventDefault: () => void}) {
    event.preventDefault();
    await alert(PromptInfo);
    setPromptInfo("");
  }

  return (
    <>
      <MenuTabs/>
      <div className="output"> Here's where the output will go</div>
      <div className="fixedBottom">
        <form onSubmit={handleSubmit}>
          <label>
            {" "}
            Enter Prompt:
            <input
              type="text"
              value={PromptInfo}
              onChange={(e) => setPromptInfo(e.target.value)} //access the current input and updates PromptInfo (e represents the event object)
            />
            <button style={{backgroundColor: "gray", color: "black"}}>
              {" "}
              Submit
            </button>
          </label>
        </form>
      </div>
    </>
  );
}
