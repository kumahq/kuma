import{d as y,B as V,J as h,c as i,o as p,a as d,w as e,h as o,b as r,g as a,i as m,j as x,O as l}from"./index-f1b8ae6a.js";import{O as G,a as C,b as B}from"./OnboardingPage-d7f0da66.js";import{B as M,F as k,G as O,e as P,h as T,A as w,_ as N,f as K}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";import{_ as S}from"./RouteTitle.vue_vue_type_script_setup_true_lang-6484968f.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-14dd845b.js";const U={class:"graph-list mb-6"},A={class:"radio-button-group"},I=y({__name:"ConfigurationTypes",setup($){const c=M(),g=k(),_={postgres:O(),memory:g,kubernetes:c},u=P(),{t:f}=T(),t=V("kubernetes");h(function(){t.value=u.getters["config/getConfigurationType"]});const v=i(()=>u.getters["config/getMulticlusterStatus"]?"onboarding-multi-zone":"onboarding-create-mesh"),b=i(()=>_[t.value]);return(z,n)=>(p(),d(N,null,{default:e(()=>[o(S,{title:r(f)("onboarding.routes.configuration-types.title")},null,8,["title"]),a(),o(w,null,{default:e(()=>[o(G,{"with-image":""},{header:e(()=>[o(C,null,{title:e(()=>[a(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[m("div",U,[(p(),d(x(b.value)))]),a(),m("div",A,[o(r(l),{modelValue:t.value,"onUpdate:modelValue":n[0]||(n[0]=s=>t.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[a(`
              Kubernetes
            `)]),_:1},8,["modelValue"]),a(),o(r(l),{modelValue:t.value,"onUpdate:modelValue":n[1]||(n[1]=s=>t.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[a(`
              Postgres
            `)]),_:1},8,["modelValue"]),a(),o(r(l),{modelValue:t.value,"onUpdate:modelValue":n[2]||(n[2]=s=>t.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[a(`
              Memory
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[o(B,{"next-step":v.value,"previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:1})]),_:1})]),_:1}))}});const q=K(I,[["__scopeId","data-v-b71119a9"]]);export{q as default};
