import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";

@Injectable({
  providedIn: 'root'
})
export class TreeService {
  constructor(private http: HttpClient) { }

  postTree(datos: any) {
    return this.http.post(`/api/agregar`, datos);
  }
}
