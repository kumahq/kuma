import{o as x}from"./kongponents.es-bdcd64fe.js";import{d as y,u as R,c,x as m,o as n,j as V,e as r,b as l,y as h,q as k,w as u,g as _,h as p,t as z,f as I,s as B,F as C}from"./index-60be81af.js";import{u as S}from"./store-94b94e8a.js";import{b as Z}from"./index-c896859e.js";import{_ as A}from"./_plugin-vue_export-helper-c27b6911.js";const T=y({__name:"ZoneIndexView",setup(D){const d=Z(),i=R(),v=S(),a=[{routeName:"zone-cp-list-view",activeRouteNames:["zone-cp-detail-view"]},{routeName:"zone-ingress-list-view",activeRouteNames:["zone-ingress-detail-view"]},{routeName:"zone-egress-list-view",activeRouteNames:["zone-egress-detail-view"]}].map(e=>({...e,title:d.t(`zones.navigation.${e.routeName}`)})),f=c(()=>a.map(e=>{const{title:s,routeName:t}=e;return{title:s,hash:"#"+t}})),N=c(()=>{const e=a.find(t=>!!(t.routeName===i.name||Array.isArray(t.activeRouteNames)&&t.activeRouteNames.includes(i.name)));return"#"+((e==null?void 0:e.routeName)??a[0].routeName)});return(e,s)=>{const t=m("router-link"),w=m("RouterView");return n(),V(C,null,[r(v).getters["config/getMulticlusterStatus"]?(n(),l(r(x),{key:0,class:"nav-tabs",tabs:f.value,"model-value":N.value},h({_:2},[k(r(a),o=>({name:`${o.routeName}-anchor`,fn:u(()=>[_(t,{to:{name:o.routeName}},{default:u(()=>[p(z(o.title),1)]),_:2},1032,["to"])])}))]),1032,["tabs","model-value"])):I("",!0),p(),_(w,null,{default:u(({Component:o,route:g})=>[(n(),l(B(o),{key:g.path}))]),_:1})],64)}}});const E=A(T,[["__scopeId","data-v-19f887ec"]]);export{E as default};
