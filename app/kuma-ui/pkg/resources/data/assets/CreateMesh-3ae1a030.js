import{d as p,h as u,r as _,i as h,o as m,j as b,w as t,a as s,e,f as o,z as g,u as n,G as f,S as v,N as y,O as x,J as S}from"./index-5f1fbf13.js";import{O as N,a as A,b as C}from"./OnboardingPage-503a3bf7.js";const r=a=>(y("data-v-bc48623a"),a=a(),x(),a),O={class:"mb-4 text-center"},k=r(()=>o("i",null,"default",-1)),B=r(()=>o("p",{class:"mt-4 text-center"},`
        This mesh is empty. Next, you add services and their data plane proxies.
      `,-1)),D=p({__name:"CreateMesh",setup(a){const c=[{label:"Name",key:"name"},{label:"Services",key:"servicesAmount"},{label:"DPPs",key:"dppsAmount"}],i=u(),d=_({total:1,data:[{name:"default",servicesAmount:0,dppsAmount:0}]}),l=h(()=>i.getters["config/getMulticlusterStatus"]?"onboarding-multi-zone":"onboarding-configuration-types");return(M,E)=>(m(),b(C,null,{header:t(()=>[s(N,null,{title:t(()=>[e(`
          Create the mesh
        `)]),_:1})]),content:t(()=>[o("p",O,[e(`
        When you install, `+g(n(f))+" creates a ",1),k,e(` mesh, but you can add as many meshes as you need.
      `)]),e(),s(n(v),{class:"table",fetcher:()=>d.value,headers:c,"disable-pagination":""},null,8,["fetcher"]),e(),B]),navigation:t(()=>[s(A,{"next-step":"onboarding-add-services","previous-step":n(l)},null,8,["previous-step"])]),_:1}))}});const T=S(D,[["__scopeId","data-v-bc48623a"]]);export{T as default};
