import{d as F,o as i,c as L,w as r,a as w,u as t,b as Z,e as u,P as O,r as e,f as A,g as H,h as J,k as V,i as X,j as D,l as s,m as Y,v as ee,F as W,n as j,t as N,p as ae,q as te,s as U,x as se,y as le,z as ne,A as S,B as oe,C as re,D as ie,E as ue}from"./index.60b0f0ac.js";import{p as ce,D as pe}from"./patchQueryParam.ae688d93.js";import{F as me}from"./FrameSkeleton.cbc6b8ea.js";import{_ as G}from"./LabelList.vue_vue_type_style_index_0_lang.bd2c37a0.js";import{T as de}from"./TabsWidget.5b63a728.js";import{_ as ve}from"./YamlView.vue_vue_type_script_setup_true_lang.152633f3.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.548da37c.js";import"./EntityStatus.6fc3c7d6.js";import"./ErrorBlock.2ee4d08e.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.be7e4bb1.js";import"./TagList.b3d2d71f.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.85e36160.js";import"./_commonjsHelpers.f037b798.js";const ye=F({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(v){const c=v;return(C,y)=>(i(),L(t(O),{class:"docs-link",appearance:"outline",target:"_blank",to:c.href},{default:r(()=>[w(t(Z),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),u(`

    Documentation
  `)]),_:1},8,["to"]))}}),fe=s("h4",null,"Dataplanes",-1),he=F({__name:"PolicyConnections",props:{mesh:{type:String,required:!0},policyType:{type:String,required:!0},policyName:{type:String,required:!0}},setup(v){const c=v,C=e(!1),y=e(!0),P=e(!1),g=e([]),f=e(""),b=A(()=>{const l=f.value.toLowerCase();return g.value.filter(({dataplane:n})=>n.name.toLowerCase().includes(l))});H(()=>c.policyName,function(){h()}),J(function(){h()});async function h(){P.value=!1,y.value=!0;try{const{items:l,total:n}=await V.getPolicyConnections({mesh:c.mesh,policyType:c.policyType,policyName:c.policyName});C.value=n>0,g.value=l!=null?l:[]}catch{P.value=!0}finally{y.value=!1}}return(l,n)=>{const E=X("router-link");return i(),D("div",null,[w(G,{"has-error":P.value,"is-loading":y.value,"is-empty":!C.value},{default:r(()=>[s("ul",null,[s("li",null,[fe,u(),Y(s("input",{id:"dataplane-search","onUpdate:modelValue":n[0]||(n[0]=p=>f.value=p),type:"text",class:"k-input mb-4",placeholder:"Filter by name",required:"","data-testid":"dataplane-search-input"},null,512),[[ee,f.value]]),u(),(i(!0),D(W,null,j(t(b),(p,k)=>(i(),D("p",{key:k,class:"my-1","data-testid":"dataplane-name"},[w(E,{to:{name:"data-plane-detail-view",params:{mesh:p.dataplane.mesh,dataPlane:p.dataplane.name}}},{default:r(()=>[u(N(p.dataplane.name),1)]),_:2},1032,["to"])]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])])}}}),_e=v=>(re("data-v-a09e01f6"),v=v(),ie(),v),ge={key:0,class:"mb-4"},be=_e(()=>s("p",null,[s("strong",null,"Warning"),u(` This policy is experimental. If you encountered any problem please open an
            `),s("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),ke={"data-testid":"policy-single-entity"},we={"data-testid":"policy-overview-tab"},De={class:"config-wrapper"},Pe=F({__name:"PolicyView",props:{policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(v){const c=v,C=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],y=ae(),P=te(),g=e(!0),f=e(!1),b=e(null),h=e(!0),l=e(!1),n=e(!1),E=e(!1),p=e({}),k=e(null),q=e(null),z=e(c.offset),I=e({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Type",key:"type"}],data:[]}),m=A(()=>P.state.policiesByPath[c.policyPath]),K=A(()=>`https://kuma.io/docs/${P.getters["config/getKumaDocsVersion"]}/policies/${m.value.path}/`);H(()=>y.params.mesh,function(){y.name===c.policyPath&&(g.value=!0,f.value=!1,h.value=!0,l.value=!1,n.value=!1,E.value=!1,b.value=null,$(0))}),$(c.offset);async function $(a){var M;z.value=a,ce("offset",a>0?a:null),g.value=!0,b.value=null;const o=y.query.ns||null,_=y.params.mesh,x=m.value.path;try{let d;if(_!==null&&o!==null)d=[await V.getSinglePolicyEntity({mesh:_,path:x,name:o})],q.value=null;else{const T={size:U,offset:a},R=await V.getAllPolicyEntitiesFromMesh({mesh:_,path:x},T);d=(M=R.items)!=null?M:[],q.value=R.next}d.length>0?(I.value.data=d.map(T=>Q(T)),E.value=!1,f.value=!1,await B({mesh:d[0].mesh,name:d[0].name,path:x})):(I.value.data=[],E.value=!0,f.value=!0,l.value=!0)}catch(d){d instanceof Error?b.value=d:console.error(d),f.value=!0}finally{g.value=!1,h.value=!1}}function Q(a){if(!a.mesh)return a;const o=a,_={name:"mesh-detail-view",params:{mesh:a.mesh}};return o.meshRoute=_,o}async function B(a){n.value=!1,h.value=!0,l.value=!1;try{const o=await V.getSinglePolicyEntity({mesh:a.mesh,path:m.value.path,name:a.name});if(o){const _=["type","name","mesh"];p.value=se(o,_),k.value=le(o)}else p.value={},l.value=!0}catch(o){n.value=!0,console.error(o)}finally{h.value=!1}}return(a,o)=>t(m)?(i(),D("div",{key:0,class:oe(["relative",t(m).path])},[t(m).isExperimental?(i(),D("div",ge,[w(t(ne),{appearance:"warning"},{alertMessage:r(()=>[be]),_:1})])):S("",!0),u(),w(me,null,{default:r(()=>[w(pe,{"selected-entity-name":p.value.name,"page-size":t(U),error:b.value,"is-loading":g.value,"empty-state":{title:"No Data",message:`There are no ${t(m).pluralDisplayName} present.`},"table-data":I.value,"table-data-is-empty":E.value,next:q.value,"page-offset":z.value,onTableAction:B,onLoadData:$},{additionalControls:r(()=>[w(ye,{href:t(K),"data-testid":"policy-documentation-link"},null,8,["href"]),u(),a.$route.query.ns?(i(),L(t(O),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:t(m).path}},{default:r(()=>[u(`
            View all
          `)]),_:1},8,["to"])):S("",!0)]),default:r(()=>[u(`
        >
        `)]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"]),u(),f.value===!1?(i(),L(de,{key:0,"has-error":b.value!==null,error:b.value,"is-loading":g.value,tabs:C,"initial-tab-override":"overview"},{tabHeader:r(()=>[s("div",null,[s("h1",ke,N(t(m).singularDisplayName)+": "+N(p.value.name),1)])]),overview:r(()=>[w(G,{"has-error":n.value,"is-loading":h.value,"is-empty":l.value},{default:r(()=>[s("div",we,[s("ul",null,[(i(!0),D(W,null,j(p.value,(_,x)=>(i(),D("li",{key:x},[s("h4",null,N(x),1),u(),s("p",null,N(_),1)]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"]),u(),s("div",De,[k.value!==null?(i(),L(ve,{key:0,id:"code-block-policy","has-error":n.value,"is-loading":h.value,"is-empty":l.value,content:k.value,"is-searchable":""},null,8,["has-error","is-loading","is-empty","content"])):S("",!0)])]),"affected-dpps":r(()=>[k.value!==null?(i(),L(he,{key:0,mesh:k.value.mesh,"policy-name":k.value.name,"policy-type":t(m).path},null,8,["mesh","policy-name","policy-type"])):S("",!0)]),_:1},8,["has-error","error","is-loading"])):S("",!0)]),_:1})],2)):S("",!0)}});const Be=ue(Pe,[["__scopeId","data-v-a09e01f6"]]);export{Be as default};
