import{d as u,B as _,c as m,o as h,a as f,w as e,h as t,b as n,g as a,i as o,t as b,A as g,y as v,z as y}from"./index-f1b8ae6a.js";import{O as x,a as A,b as S}from"./OnboardingPage-d7f0da66.js";import{e as N,h as C,A as k,_ as B,f as I}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";import{_ as w}from"./RouteTitle.vue_vue_type_script_setup_true_lang-6484968f.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-14dd845b.js";const i=s=>(v("data-v-94485eb5"),s=s(),y(),s),D={class:"mb-4 text-center"},M=i(()=>o("i",null,"default",-1)),O=i(()=>o("p",{class:"mt-4 text-center"},`
            This mesh is empty. Next, you add services and their data plane proxies.
          `,-1)),V=u({__name:"CreateMesh",setup(s){const c=[{label:"Name",key:"name"},{label:"Services",key:"servicesAmount"},{label:"DPPs",key:"dppsAmount"}],l=N(),{t:r}=C(),d=_({total:1,data:[{name:"default",servicesAmount:0,dppsAmount:0}]}),p=m(()=>l.getters["config/getMulticlusterStatus"]?"onboarding-multi-zone":"onboarding-configuration-types");return(E,P)=>(h(),f(B,null,{default:e(()=>[t(w,{title:n(r)("onboarding.routes.create-mesh.title")},null,8,["title"]),a(),t(k,null,{default:e(()=>[t(x,null,{header:e(()=>[t(A,null,{title:e(()=>[a(`
              Create the mesh
            `)]),_:1})]),content:e(()=>[o("p",D,[a(`
            When you install, `+b(n(r)("common.product.name"))+" creates a ",1),M,a(` mesh, but you can add as many meshes as you need.
          `)]),a(),t(n(g),{class:"table",fetcher:()=>d.value,headers:c,"disable-pagination":""},null,8,["fetcher"]),a(),O]),navigation:e(()=>[t(S,{"next-step":"onboarding-add-services","previous-step":p.value},null,8,["previous-step"])]),_:1})]),_:1})]),_:1}))}});const R=I(V,[["__scopeId","data-v-94485eb5"]]);export{R as default};
