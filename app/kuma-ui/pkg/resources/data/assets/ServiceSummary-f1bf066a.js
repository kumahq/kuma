import{F as w}from"./kongponents.es-a62ec79b.js";import{D as c,a as B}from"./DefinitionListItem-5764f62f.js";import{S as P}from"./StatusBadge-f4a1cf45.js";import{T as C}from"./TagList-2f6ed4b9.js";import{_ as F}from"./YamlView.vue_vue_type_script_setup_true_lang-f5b58b07.js";import{d as L,c as a,K as N,x as V,e as o,w as r,b as s,o as n,i as v,h as t,g as i,t as u,f as m,j as f,F as k}from"./index-5550fdb4.js";import{_ as j}from"./_plugin-vue_export-helper-c27b6911.js";const O={class:"entity-section-list"},$={class:"entity-title"},q={key:0,class:"config-section"},E=L({__name:"ServiceSummary",props:{service:{type:Object,required:!0},externalService:{type:Object,required:!1,default:null}},setup(h){const e=h,b=a(()=>({name:"service-detail-view",params:{service:e.service.name,mesh:e.service.mesh}})),p=a(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.networking.address:e.service.addressPort??null),x=a(()=>{var l;return e.service.serviceType==="external"&&e.externalService!==null?(l=e.externalService.networking.tls)!=null&&l.enabled?"Enabled":"Disabled":null}),y=a(()=>{var l,d;if(e.service.serviceType==="external")return null;{const _=((l=e.service.dataplanes)==null?void 0:l.online)??0,D=((d=e.service.dataplanes)==null?void 0:d.total)??0;return`${_} online / ${D} total`}}),S=a(()=>e.service.serviceType==="external"?null:e.service.status??null),g=a(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.tags:null),T=a(()=>N(e.externalService??e.service));return(l,d)=>{const _=V("router-link");return n(),o(s(w),null,{body:r(()=>[v("div",O,[v("section",null,[v("h1",$,[v("span",null,[t(`
              Service:

              `),i(_,{to:s(b)},{default:r(()=>[t(u(e.service.name),1)]),_:1},8,["to"])]),t(),s(S)?(n(),o(P,{key:0,status:s(S)},null,8,["status"])):m("",!0)]),t(),i(B,{class:"mt-4"},{default:r(()=>[i(c,{term:"Mesh"},{default:r(()=>[t(u(e.service.mesh),1)]),_:1}),t(),i(c,{term:"Address"},{default:r(()=>[s(p)!==null?(n(),f(k,{key:0},[t(u(s(p)),1)],64)):(n(),f(k,{key:1},[t(`
                —
              `)],64))]),_:1}),t(),s(x)!==null?(n(),o(c,{key:0,term:"TLS"},{default:r(()=>[t(u(s(x)),1)]),_:1})):m("",!0),t(),s(y)!==null?(n(),o(c,{key:1,term:"Data Plane Proxies"},{default:r(()=>[t(u(s(y)),1)]),_:1})):m("",!0),t(),s(g)!==null?(n(),o(c,{key:2,term:"Tags"},{default:r(()=>[i(C,{tags:s(g)},null,8,["tags"])]),_:1})):m("",!0)]),_:1})]),t(),e.service.serviceType==="external"?(n(),f("section",q,[i(F,{id:"code-block-service",content:s(T),"is-searchable":"","code-max-height":"250px"},null,8,["content"])])):m("",!0)])]),_:1})}}});const H=j(E,[["__scopeId","data-v-fbaa2f3f"]]);export{H as S};
