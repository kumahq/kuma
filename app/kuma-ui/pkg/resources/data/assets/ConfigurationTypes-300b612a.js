import{n as u}from"./kongponents.es-3ba46133.js";import{O as y,a as b,b as V}from"./OnboardingPage-92514748.js";import{e as h,f as x,g as C}from"./index-319041fc.js";import{u as G}from"./store-f2021894.js";import{d as M,r as k,i as P,c as p,o as m,a as i,w as o,e as s,f as t,g as d,m as T,u as r}from"./runtime-dom.esm-bundler-9284044f.js";import{_ as B}from"./_plugin-vue_export-helper-c27b6911.js";import"./production-554ae9d4.js";import"./index-36b3783c.js";import"./datadogLogEvents-302eea7b.js";import"./DoughnutChart-861842b2.js";const N={class:"graph-list mb-6"},O={class:"radio-button-group"},w=M({__name:"ConfigurationTypes",setup(K){const g=h(),c=x(),f={postgres:C(),memory:c,kubernetes:g},l=G(),e=k("kubernetes");P(function(){e.value=l.getters["config/getConfigurationType"]});const _=p(()=>l.getters["config/getMulticlusterStatus"]?"onboarding-multi-zone":"onboarding-create-mesh"),v=p(()=>f[e.value]);return(U,n)=>(m(),i(V,{"with-image":""},{header:o(()=>[s(y,null,{title:o(()=>[t(`
          Learn about configuration storage
        `)]),_:1})]),content:o(()=>[d("div",N,[(m(),i(T(r(v))))]),t(),d("div",O,[s(r(u),{modelValue:e.value,"onUpdate:modelValue":n[0]||(n[0]=a=>e.value=a),name:"deployment","selected-value":"kubernetes"},{default:o(()=>[t(`
          Kubernetes
        `)]),_:1},8,["modelValue"]),t(),s(r(u),{modelValue:e.value,"onUpdate:modelValue":n[1]||(n[1]=a=>e.value=a),name:"deployment","selected-value":"postgres"},{default:o(()=>[t(`
          Postgres
        `)]),_:1},8,["modelValue"]),t(),s(r(u),{modelValue:e.value,"onUpdate:modelValue":n[2]||(n[2]=a=>e.value=a),name:"deployment","selected-value":"memory"},{default:o(()=>[t(`
          Memory
        `)]),_:1},8,["modelValue"])])]),navigation:o(()=>[s(b,{"next-step":r(_),"previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:1}))}});const J=B(w,[["__scopeId","data-v-673391df"]]);export{J as default};
