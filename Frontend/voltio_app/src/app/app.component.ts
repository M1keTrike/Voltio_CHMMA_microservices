import { Component } from '@angular/core';


@Component({
  selector: 'app-root',

  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  title = 'voltio_app';
  holdInterval: any;

  counter: number = 0;

  increment() {
    this.counter++;
  }

  decrement() {
    if (this.counter > 0) this.counter--;
  }

  startHolding(action: 'increment' | 'decrement') {
    this.holdInterval = setInterval(() => {
      if (action === 'increment') {
        this.increment();
      } else {
        this.decrement();
      }
    }, 100);
  }

  stopHolding() {
    clearInterval(this.holdInterval);
  }

  mouseUp() {
    this.stopHolding();
  }
}
