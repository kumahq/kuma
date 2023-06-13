import{l as F,T as $,c as O,b as R,n as D,d as L}from"./kongponents.es-0bcabadf.js";import{d as Z,q as B,m as y,o as f,f as C,e as o,t,u as n,b as e,a as l,p as W,j as M,y as N,r as b,c as I,w as r,g as S,F as j,R as H}from"./index-95f89edf.js";import{_ as z}from"./CodeBlock.vue_vue_type_style_index_0_lang-10589c0c.js";import{a as G,g as q,b as x,_ as T,u as J,f as Q,e as X}from"./RouteView.vue_vue_type_script_setup_true_lang-9e79b18b.js";import{_ as Y}from"./RouteTitle.vue_vue_type_script_setup_true_lang-9b936e20.js";import{E as ee}from"./ErrorBlock-a4644cb1.js";import{_ as ne}from"./EntityScanner.vue_vue_type_script_setup_true_lang-a871648a.js";const oe=m=>(W("data-v-293c555d"),m=m(),M(),m),te={href:"https://helm.sh/docs/intro/install/"},se=oe(()=>o("p",null,"On your local machine, create a namespace in your Kubernetes cluster and pull down the kong Helm repo.",-1)),ae={class:"k-input-label mt-4"},le={class:"mt-4"},re=Z({__name:"ZoneCreateKubernetesInstructions",props:{zoneName:{type:String,required:!0},zoneIngressEnabled:{type:Boolean,required:!0},zoneEgressEnabled:{type:Boolean,required:!0},token:{type:String,required:!0},base64EncodedToken:{type:String,required:!0}},setup(m){const a=m,u=G(),s=q(),p=B(),k=x(),g=y(()=>s.t("zones.form.kubernetes.secret.createSecretCommand",{token:a.base64EncodedToken}).trim()),_=y(()=>{const i={zoneName:a.zoneName,globalKdsAddress:k.state.globalKdsAddress,zoneIngressEnabled:String(a.zoneIngressEnabled),zoneEgressEnabled:String(a.zoneEgressEnabled)};return typeof p.params.virtualControlPlaneId=="string"&&(i.controlPlaneId=p.params.virtualControlPlaneId),s.t("zones.form.kubernetes.connectZone.config",i).trim()});return(i,h)=>(f(),C("div",null,[o("h3",null,"1. "+t(n(s).t("zones.form.kubernetes.prerequisites.title")),1),e(),o("ul",null,[o("li",null,[o("b",null,t(n(s).t("zones.form.kubernetes.prerequisites.step1Label"))+t(a.zoneIngressEnabled?" "+n(s).t("zones.form.kubernetes.prerequisites.step1LabelAddendum"):""),1),e(`:
        `+t(n(s).t("zones.form.kubernetes.prerequisites.step1Description",{productName:n(u)("KUMA_PRODUCT_NAME")})),1)]),e(),o("li",null,[o("b",null,t(n(s).t("zones.form.kubernetes.prerequisites.step2Label")),1),e(`:
        `+t(n(s).t("zones.form.kubernetes.prerequisites.step2Description")),1)]),e(),o("li",null,[o("a",te,t(n(s).t("zones.form.kubernetes.prerequisites.step3LinkTitle")),1),e(" "+t(n(s).t("zones.form.kubernetes.prerequisites.step3Tail")),1)])]),e(),o("h3",null,"2. "+t(n(s).t("zones.form.kubernetes.helm.title")),1),e(),se,e(),o("ol",null,[o("li",null,[e(t(n(s).t("zones.form.kubernetes.helm.step1Description"))+" ",1),l(z,{id:"zone-kubernetes-create-namespace",class:"mt-4",code:n(s).t("zones.form.kubernetes.helm.step1Command"),language:"bash"},null,8,["code"])]),e(),o("li",null,[e(t(n(s).t("zones.form.kubernetes.helm.step2Description"))+" ",1),l(z,{id:"zone-kubernetes-add-charts-repo",class:"mt-4",code:n(s).t("zones.form.kubernetes.helm.step2Command"),language:"bash"},null,8,["code"])]),e(),o("li",null,[e(t(n(s).t("zones.form.kubernetes.helm.step3Description"))+" ",1),l(z,{id:"zone-kubernetes-repo-update",class:"mt-4",code:n(s).t("zones.form.kubernetes.helm.step3Command"),language:"bash"},null,8,["code"])])]),e(),o("h3",null,"3. "+t(n(s).t("zones.form.kubernetes.secret.title")),1),e(),o("p",null,t(n(s).t("zones.form.kubernetes.secret.createSecretDescription")),1),e(),l(z,{id:"zone-kubernetes-create-secret",class:"mt-4",code:g.value,language:"bash"},null,8,["code"]),e(),o("h3",null,"4. "+t(n(s).t("zones.form.kubernetes.connectZone.title")),1),e(),o("p",null,t(n(s).t("zones.form.kubernetes.connectZone.configDescription")),1),e(),o("span",ae,t(n(s).t("zones.form.kubernetes.connectZone.configFileName")),1),e(),l(z,{id:"zone-kubernetes-config-code-block",code:_.value,language:"yaml"},null,8,["code"]),e(),o("p",le,t(n(s).t("zones.form.kubernetes.connectZone.connectDescription")),1),e(),l(z,{id:"zone-kubernetes-command-code-block",class:"mt-4",code:n(s).t("zones.form.kubernetes.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}});const ue=T(re,[["__scopeId","data-v-293c555d"]]),ie={class:"k-input-label mt-4"},ce={class:"mt-4"},de=Z({__name:"ZoneCreateUniversalInstructions",props:{zoneName:{type:String,required:!0},token:{type:String,required:!0}},setup(m){const a=m,u=q(),s=B(),p=x(),k=y(()=>u.t("zones.form.universal.saveToken.saveTokenCommand",{token:a.token}).trim()),g=y(()=>{const _={zoneName:a.zoneName,globalKdsAddress:p.state.globalKdsAddress};return typeof s.params.virtualControlPlaneId=="string"&&(_.controlPlaneId=s.params.virtualControlPlaneId),u.t("zones.form.universal.connectZone.config",_).trim()});return(_,i)=>(f(),C("div",null,[o("h3",null,"1. "+t(n(u).t("zones.form.universal.saveToken.title")),1),e(),o("p",null,t(n(u).t("zones.form.universal.saveToken.saveTokenDescription")),1),e(),l(z,{id:"zone-kubernetes-token",class:"mt-4",code:k.value,language:"bash"},null,8,["code"]),e(),o("h3",null,"2. "+t(n(u).t("zones.form.universal.connectZone.title")),1),e(),o("p",null,t(n(u).t("zones.form.universal.connectZone.configDescription")),1),e(),o("span",ie,t(n(u).t("zones.form.universal.connectZone.configFileName")),1),e(),l(z,{id:"zone-universal-config-code-block",class:"mt-4",code:g.value,language:"yaml"},null,8,["code"]),e(),o("p",ce,t(n(u).t("zones.form.universal.connectZone.connectDescription")),1),e(),l(z,{id:"zone-universal-connect-command-code-block",class:"mt-4",code:n(u).t("zones.form.universal.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}});const me=T(de,[["__scopeId","data-v-ea760a7e"]]),pe={class:"app-title-bar"},_e={class:"title-wrapper"},ve={class:"title"},be={class:"action-list"},fe=Z({__name:"WizardTitleBar",setup(m){return(a,u)=>(f(),C("div",pe,[o("div",_e,[l(n(F),{icon:"kong"}),e(),o("span",ve,[N(a.$slots,"title",{},void 0,!0)])]),e(),o("div",be,[N(a.$slots,"actions",{},void 0,!0)])]))}});const ze=T(fe,[["__scopeId","data-v-f0a75c34"]]),ke={class:"form-content"},ge={class:"form-wrapper mt-4"},he={key:1,class:"form-wrapper mt-4"},ye={class:"k-input-label"},Ce={class:"radio-button-group"},Ee={class:"k-input-label"},Ie={class:"radio-button-group"},Ze={class:"k-input-label"},Te={class:"radio-button-group"},Ve={class:"mt-6"},we={class:"mt-2"},$e=Z({__name:"ZoneCreateView",setup(m){const{t:a}=q(),u=J(),s=b(null),p=b(!1),k=b(null),g=b(!1),_=b(null),i=b(""),h=b("kubernetes"),V=b(!0),w=b(!0),E=y(()=>s.value!==null&&s.value.token?s.value.token:""),K=y(()=>E.value!==""?window.btoa(E.value):""),U=y(()=>i.value!=="");async function A(){p.value=!0,k.value=null;try{s.value=await u.createZone({name:i.value})}catch(d){d instanceof Error?k.value=d:console.error(d)}finally{p.value=!1}}async function P(){g.value=!1,_.value=null;try{const d=await u.getZoneOverview({name:i.value}),c=H(d.zoneInsight);g.value=c==="online"}catch(d){d instanceof Error?_.value=d:console.error(d)}}return(d,c)=>(f(),I(X,null,{default:r(()=>[l(Y,{title:n(a)("zones.routes.create.title")},null,8,["title"]),e(),l(Q,{breadcrumbs:[]},{default:r(()=>[l(ze,{class:"mb-6"},{title:r(()=>[e(t(n(a)("zones.routes.create.title")),1)]),actions:r(()=>[l(n($),{appearance:"outline",to:{name:"zone-cp-list-view"}},{default:r(()=>[e(t(n(a)("zones.form.exit")),1)]),_:1})]),_:1}),e(),o("div",ke,[o("h1",null,t(n(a)("zones.routes.create.title")),1),e(),o("div",ge,[o("div",null,[l(n(O),{for:"zone-name"},{default:r(()=>[e(t(n(a)("zones.form.nameLabel"))+` *
            `,1)]),_:1}),e(),l(n(R),{id:"zone-name",modelValue:i.value,"onUpdate:modelValue":c[0]||(c[0]=v=>i.value=v),type:"text",name:"zone-name","data-testid":"name-input",disabled:s.value!==null},null,8,["modelValue","disabled"])]),e(),l(n($),{appearance:"creation",icon:p.value?"spinner":"plus",disabled:!U.value||p.value||s.value!==null,"data-testid":"create-zone-button",onClick:A},{default:r(()=>[e(t(n(a)("zones.form.createZoneButtonLabel")),1)]),_:1},8,["icon","disabled"])]),e(),k.value!==null?(f(),I(ee,{key:0,class:"mt-4",error:k.value},{default:r(()=>[e(t(n(a)("zones.create.errorTitle")),1)]),_:1},8,["error"])):S("",!0),e(),s.value!==null?(f(),C("div",he,[o("div",null,[o("span",ye,t(n(a)("zones.form.environmentLabel"))+` *
            `,1),e(),o("div",Ce,[l(n(D),{id:"zone-environment-universal",modelValue:h.value,"onUpdate:modelValue":c[1]||(c[1]=v=>h.value=v),"selected-value":"universal",name:"zone-environment","data-testid":"environment-universal-radio-button"},{default:r(()=>[e(t(n(a)("zones.form.universalLabel")),1)]),_:1},8,["modelValue"]),e(),l(n(D),{id:"zone-environment-kubernetes",modelValue:h.value,"onUpdate:modelValue":c[2]||(c[2]=v=>h.value=v),"selected-value":"kubernetes",name:"zone-environment","data-testid":"environment-kubernetes-radio-button"},{default:r(()=>[e(t(n(a)("zones.form.kubernetesLabel")),1)]),_:1},8,["modelValue"])])]),e(),h.value==="kubernetes"?(f(),C(j,{key:0},[o("div",null,[o("span",Ee,t(n(a)("zones.form.zoneIngressLabel"))+` *
              `,1),e(),o("div",Ie,[l(n(L),{id:"zone-ingress-enabled",modelValue:V.value,"onUpdate:modelValue":c[3]||(c[3]=v=>V.value=v),"data-testid":"ingress-input-switch"},{label:r(()=>[e(t(n(a)("zones.form.zoneIngressEnabledLabel")),1)]),_:1},8,["modelValue"])])]),e(),o("div",null,[o("span",Ze,t(n(a)("zones.form.zoneEgressLabel"))+` *
              `,1),e(),o("div",Te,[l(n(L),{id:"zone-egress-enabled",modelValue:w.value,"onUpdate:modelValue":c[4]||(c[4]=v=>w.value=v),"data-testid":"egress-input-switch"},{label:r(()=>[e(t(n(a)("zones.form.zoneEgressEnabledLabel")),1)]),_:1},8,["modelValue"])])])],64)):S("",!0),e(),o("h2",Ve,t(n(a)("zones.form.connectZone")),1),e(),h.value==="universal"?(f(),I(me,{key:1,"zone-name":i.value,token:E.value},null,8,["zone-name","token"])):(f(),I(ue,{key:2,"zone-name":i.value,"zone-ingress-enabled":V.value,"zone-egress-enabled":w.value,token:E.value,"base64-encoded-token":K.value},null,8,["zone-name","zone-ingress-enabled","zone-egress-enabled","token","base64-encoded-token"])),e(),l(ne,{"loader-function":P,"has-error":_.value!==null,"can-complete":g.value},{"loading-title":r(()=>[e(t(n(a)("zones.form.scan.waitTitle")),1)]),"complete-title":r(()=>[e(t(n(a)("zones.form.scan.completeTitle")),1)]),"complete-content":r(()=>[o("p",null,t(n(a)("zones.form.scan.completeDescription",{name:i.value})),1),e(),o("p",we,[l(n($),{appearance:"primary",to:{name:"zone-cp-detail-view",params:{zone:i.value}}},{default:r(()=>[e(t(n(a)("zones.form.scan.completeButtonLabel",{name:i.value})),1)]),_:1},8,["to"])])]),"error-title":r(()=>[o("h3",null,t(n(a)("zones.form.scan.errorTitle")),1)]),"error-content":r(()=>[o("p",null,t(n(a)("zones.form.scan.errorDescription")),1)]),_:1},8,["has-error","can-complete"])])):S("",!0)])]),_:1})]),_:1}))}});const Ke=T($e,[["__scopeId","data-v-1754afd6"]]);export{Ke as default};
