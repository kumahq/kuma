import{d as V,l as h,c as C,o as l,a as i,w as e,h as a,b as r,g as t,e as x,i as m,j as G,z as u,F as k,f as B}from"./index-9a3d231d.js";import{O as N,a as P,b as w}from"./OnboardingPage-e007feea.js";import{l as M,m as O,n as T,g as z,i as K,A as U,_ as $,f as A}from"./RouteView.vue_vue_type_script_setup_true_lang-da83f5a8.js";import{_ as F}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3a51c48f.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-fe937ad6.js";const I={class:"graph-list mb-6"},j={class:"radio-button-group"},D=V({__name:"ConfigurationTypes",setup(E){const p=M(),c=O(),_={postgres:T(),memory:c,kubernetes:p},{t:g}=z(),o=h("kubernetes"),f=d=>{o.value=d.store.type},v=C(()=>_[o.value]);return(d,n)=>(l(),i($,null,{default:e(({can:b})=>[a(F,{title:r(g)("onboarding.routes.configuration-types.title")},null,8,["title"]),t(),a(U,null,{default:e(()=>[a(N,{"with-image":""},{header:e(()=>[a(P,null,{title:e(()=>[t(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[a(K,{src:"/config",onChange:f},{default:e(({data:y})=>[typeof y<"u"?(l(),x(k,{key:0},[m("div",I,[(l(),i(G(v.value)))]),t(),m("div",j,[a(r(u),{modelValue:o.value,"onUpdate:modelValue":n[0]||(n[0]=s=>o.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[t(`
                  Kubernetes
                `)]),_:1},8,["modelValue"]),t(),a(r(u),{modelValue:o.value,"onUpdate:modelValue":n[1]||(n[1]=s=>o.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[t(`
                  Postgres
                `)]),_:1},8,["modelValue"]),t(),a(r(u),{modelValue:o.value,"onUpdate:modelValue":n[2]||(n[2]=s=>o.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[t(`
                  Memory
                `)]),_:1},8,["modelValue"])])],64)):B("",!0)]),_:1})]),navigation:e(()=>[a(w,{"next-step":b("use zones")?"onboarding-multi-zone":"onboarding-create-mesh","previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1}))}});const S=A(D,[["__scopeId","data-v-5db0c53c"]]);export{S as default};
