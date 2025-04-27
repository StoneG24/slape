import "./App.css";
interface dropDownSettings {
  className: string;
  value: string;
  callBack: (event: string) => void;
  optionObject: {name: string; type: string | number}[];
}

const themeColor = localStorage.getItem("StyleSetting");

export default function DropDownButton({
  className,
  value,
  callBack,
  optionObject,
}: dropDownSettings) {
  return (
    <select
      className={`${themeColor}_${className}`}
      value={value}
      onChange={(event) => {
        callBack(event.target.value);
      }}
    >
      {optionObject.map((type) => {
        return (
          <option value={type.type} key={type.type}>
            {type.name}
          </option>
        );
      })}
    </select>
  );
}
