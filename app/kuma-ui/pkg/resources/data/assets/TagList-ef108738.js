import{d as f,e as b,a8 as g,f as h,o as r,j as l,F as y,G as k,g as i,w as p,n as x,l as L,D as d,m as w,i as T,ah as B,q as j}from"./index-1119a66f.js";function z(o){return Object.entries(o??{}).map(([s,a])=>({label:s,value:a}))}const C={class:"tag-list"},D=f({__name:"TagList",props:{tags:{type:Object,required:!0}},setup(o){const s=o,a=b(),c=g(),m=h(()=>(Array.isArray(s.tags)?s.tags:z(s.tags)).map(n=>{const{label:t,value:u}=n,v=_(n);return{label:t,value:u,route:v}}));function _(e){if(e.value!=="*")try{switch(e.label){case"kuma.io/zone":return c.resolve({name:"zone-cp-detail-view",params:{zone:e.value}});case"kuma.io/service":return"mesh"in a.params?c.resolve({name:"service-detail-view",params:{mesh:a.params.mesh,service:e.value}}):void 0;default:return}}catch{return}}return(e,n)=>(r(),l("span",C,[(r(!0),l(y,null,k(m.value,(t,u)=>(r(),i(T(B),{key:u,class:"tag-badge"},{default:p(()=>[(r(),i(x(t.route?"router-link":"span"),{to:t.route},{default:p(()=>[L(d(t.label)+":",1),w("b",null,d(t.value),1)]),_:2},1032,["to"]))]),_:2},1024))),128))]))}});const q=j(D,[["__scopeId","data-v-94e5d380"]]);export{q as T};
