import "./pipelineCard.css";
import {useState} from "react";
import Modal from "./Modal.tsx";
import {createPortal} from "react-dom";
import DropDownButton from "./DropDownButton.tsx";
import {ReactNode} from "react";

interface pipelineProperties {
  pipeline: string;
  displayName: string;
  description: string;
}

export default function PipelineCard({
  pipeline,
  displayName,
  description,
}: pipelineProperties) {
  const [ModalOpen, setModalOpen] = useState(false);
  const [modelName, setmodelName] = useState("Phi 3");
  const [ ModelList, setModelList ] = useState([]);
  const [Models, setModels] = useState<object[]>([]);
  const [ AddModelsButtonStatus, setAddModelsButtonStatus ] = useState("addModelActive");

  if (localStorage.getItem("PromptSetting") == null)
    localStorage.setItem("PromptSetting", "Automatic");
  if (localStorage.getItem("StyleSetting") == null)
    localStorage.setItem("StyleSetting", "Dark");

  const themeColor: string | null = localStorage.getItem("StyleSetting");

  const pipelineSettingsButtonHandler = () => {
    setModalOpen(true);
    getModelsFromVito();
    setModels([]);
    localStorage.removeItem(`${pipeline}Models`);
    setAddModelsButtonStatus("addModelActive");
  };

  const modalCloseButtonHandler = () => {
    setModalOpen(false);
    localStorage.setItem(`${pipeline}Models`, JSON.stringify(Models));
  };

  async function getModelsFromVito() {
    const dropDownOptions: {type: string, name: string}[] = [];
    const response = await fetch(`http://localhost:8080/getmodels`, {
      method: "GET",
    });

    let responseBody: {models: string[]};
    let models: string[] = [];

    if (response.ok) {
      responseBody = await response.json();
      models = responseBody.models;
    } else {
      alert("Failed to pull models. Please restart front end and ensure backend is running.")
    }

    models.forEach((element) => {
      dropDownOptions.push({type: element, name:element.slice(0, -5)})
    });

    setModelList(dropDownOptions);
    setmodelName(dropDownOptions[0].name);
  }
  // function addModelHandler() {
  //   if (localStorage.getItem(`${pipeline}Models`) == null)
  //     localStorage.setItem(`${pipeline}Models`, JSON.stringify([modelName]));
  //   else {
  //     const currentPipelineModel: string[] = JSON.parse(localStorage.getItem(`${pipeline}Models`) as string);
  //     currentPipelineModel.push(modelName);
  //     localStorage.setItem(
  //       `${pipeline}Models`,
  //       JSON.stringify(currentPipelineModel)
  //     );
  //   }
  // }

  function addModelHandler() {
    console.log(modelName);
    const modelObject = {name: modelName.slice(0, -5), fullName: modelName};
    

    if (pipeline !== "simple") {
      setModels([...Models, modelObject]);
    } else if (Models.length === 0) {
      setModels([modelObject]);
      setAddModelsButtonStatus("addModelInactive");
    }
  }

  function displayCurrentModels(className: string): ReactNode {
    let modelsString = "";
    Models.forEach(
      (element: {name: string; fullName: string}, index: number) => {
        if (index !== Models.length - 1) {
          modelsString += element.name + ", ";
        } else {
          modelsString += element.name;
        }
      }
    );
    return <p className={className}>{`Current Models: ${modelsString}`}</p>;
  }

  return (
    <div className={`${themeColor}_pipelineDiv`} tabIndex={0}>
      <h3 className="pipelineHeader">{displayName}</h3>
      <button
        className={`${themeColor}_pipelineButton`}
        onClick={pipelineSettingsButtonHandler}
      >
        Settings
      </button>
      {createPortal(
        <Modal isOpen={ModalOpen} onClose={modalCloseButtonHandler}>
          <p>
            {/* {displayCurrentModels("")} */}
            <button
              className={`${themeColor}_${AddModelsButtonStatus}`}
              onClick={addModelHandler}
            >
              Add Model
            </button>
            <DropDownButton
              className="modelDropdown"
              value={modelName}
              callBack={setmodelName}
              optionObject={ModelList}
            />
          </p>
        </Modal>,
        document.body
      )}
      {displayCurrentModels("pipelineModels")}
      <p className="pipelineDesc">{`Description: ${description}`}</p>
    </div>
  );
}
