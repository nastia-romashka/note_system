export function FlashMessage({ message, onClose }) {
  return (
    <div className={`flash-message ${message.type}`}>
      <span>{message.text}</span>
      <button type="button" onClick={onClose}>
        закрыть
      </button>
    </div>
  );
}
